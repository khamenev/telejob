package process

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/zelenin/go-tdlib/client"
	"io/ioutil"
	"log"
	"telejob/internal/config"
	"telejob/internal/db"
	"time"
)

// LastMessageIDs Структура для хранения последних ID сообщений по каналам
type LastMessageIDs struct {
	IDs map[string]int64 `json:"ids"`
}

func ProcessChannels(cfg *config.Config, tdClient *client.Client, database *sql.DB) {
	const lastIDsFilename = "lastMessageIDs.json"

	lastIDs, err := loadLastMessageIDs(lastIDsFilename)
	if err != nil {
		log.Printf("Error loading last message IDs: %v", err)
	}

	for _, channelName := range cfg.Channels {
		log.Printf("<-- Start iteration for %s", channelName)
		lastMessageID := getLastMessageID(lastIDs, channelName)

		channelData, err := getChannelDataOrResolve(tdClient, database, channelName)
		if err != nil {
			log.Printf("Error getting or resolving channel data for %s: %v", channelName, err)
			continue
		}

		filteredMessages, newLastMessageID := processChatHistory(tdClient, channelData, lastMessageID)
		lastIDs.IDs[channelName] = newLastMessageID

		saveLastMessageIDs(lastIDsFilename, lastIDs)

		if len(filteredMessages) > 0 {
			err = forwardMessages(context.Background(), tdClient, filteredMessages, channelData.ReceiverID)
			if err != nil {
				log.Printf("Error forwarding messages for channel %s: %v", channelName, err)
			}
		} else {
			log.Printf("No messages to forward for %s", channelName)
		}

		log.Printf("<-- End iteration for %s", channelName)
		time.Sleep(1 * time.Second)
	}
}

func getLastMessageID(lastIDs LastMessageIDs, channelName string) int64 {
	if lastMessageID, ok := lastIDs.IDs[channelName]; ok {
		return lastMessageID
	}
	return 0
}

func getChannelDataOrResolve(tdClient *client.Client, database *sql.DB, channelName string) (db.ChannelData, error) {
	channelData, err := db.GetChannelData(database, channelName)
	if err != nil {
		return db.ChannelData{}, err
	}

	if channelData.ChannelID <= 0 {
		channel, err := tdClient.SearchPublicChat(&client.SearchPublicChatRequest{Username: channelName})
		if err != nil {
			return db.ChannelData{}, err
		}
		channelData.ChannelID = channel.Id
		err = db.SaveOrUpdateChannelData(database, channelData)
		if err != nil {
			return db.ChannelData{}, err
		}
	}

	return channelData, nil
}

func processChatHistory(tdClient *client.Client, channelData db.ChannelData, lastMessageID int64) ([]Message, int64) {
	var filteredMessages []Message
	var offset, limit = -10, 10
	var maxID int64
	log.Printf("lastMessageID before iteration: %v", lastMessageID)

	for {
		request := &client.GetChatHistoryRequest{
			ChatId:        channelData.ChannelID,
			FromMessageId: lastMessageID,
			Offset:        int32(offset),
			Limit:         int32(limit),
			OnlyLocal:     false,
		}

		time.Sleep(1 * time.Second)

		messages, err := tdClient.GetChatHistory(request)
		if err != nil || len(messages.Messages) == 0 {
			log.Printf("err != nil || len(messages.Messages) == 0 ")
			break
		}

		for _, message := range messages.Messages {
			if message.Id > maxID {
				maxID = message.Id
			}
		}

		messageFilter := NewMessageFilter(channelData.IncludeKeywords, channelData.ExcludeKeywords)
		filteredMessages = append(filteredMessages, convertAndFilterMessages(lastMessageID, messages.Messages, messageFilter)...)
		lastMessageID = maxID
		if maxID == 0 || len(messages.Messages) < limit {
			log.Printf("maxID == 0 || len(messages.Messages) < limit")
			break
		}
		log.Printf("lastMessageID after iteration: %v", lastMessageID)
	}

	return filteredMessages, maxID
}

func forwardMessages(ctx context.Context, tdClient *client.Client, messages []Message, receiverUsername string) error {
	receiver, err := tdClient.SearchPublicChat(&client.SearchPublicChatRequest{Username: receiverUsername})
	if err != nil {
		return err
	}

	return forwardFilteredMessages(ctx, tdClient, messages, receiver.Id)
}

func convertAndFilterMessages(lastMessageID int64, msgs []*client.Message, filter *MessageFilter) []Message {
	var internalMessages []Message
	for _, message := range msgs {
		switch m := message.Content.(type) {
		case *client.MessageDocument:
			continue
		case *client.MessageText:
			internalMessage := Message{
				ID:        message.Id,
				Text:      m.Text.Text,
				ChannelId: message.ChatId,
			}
			if internalMessage.ID > lastMessageID {
				internalMessages = append(internalMessages, internalMessage)
			}
		default:
			continue
		}
	}

	return filter.Filter(internalMessages)
}

func forwardFilteredMessages(ctx context.Context, tdClient *client.Client, messages []Message, receiverID int64) error {
	for _, msg := range messages {
		messageIds := []int64{msg.ID}
		request := &client.ForwardMessagesRequest{
			ChatId:     receiverID,
			FromChatId: msg.ChannelId,
			MessageIds: messageIds,
		}
		log.Printf("Forwarding message ID: %v", msg.ID)

		_, err := tdClient.ForwardMessages(request)
		if err != nil {
			log.Printf("Failed to forward message %d from chat %d: %v", msg.ID, msg.ChannelId, err)
			continue
		}
		log.Printf("Forwarded message ID: %v", msg.ID)

		select {
		case <-time.After(1 * time.Second):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func saveLastMessageIDs(filename string, lastIDs LastMessageIDs) error {
	data, err := json.Marshal(lastIDs)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func loadLastMessageIDs(filename string) (LastMessageIDs, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return LastMessageIDs{IDs: make(map[string]int64)}, err
	}
	var lastIDs LastMessageIDs
	if err := json.Unmarshal(data, &lastIDs); err != nil {
		return LastMessageIDs{IDs: make(map[string]int64)}, err
	}
	return lastIDs, nil
}
