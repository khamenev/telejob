package db

import (
	"database/sql"
	"encoding/json"
)

type ChannelData struct {
	Name            string
	ChannelID       int64
	AccessHash      int64
	ReceiverID      string
	IncludeKeywords []string
	ExcludeKeywords []string
}

// Получение данных канала из БД
func GetChannelData(db *sql.DB, channelName string) (ChannelData, error) {
	var data ChannelData
	var includeKeywordsJSON []byte // Используем переменную для временного хранения JSON из БД
	var excludeKeywordsJSON []byte

	err := db.QueryRow("SELECT name, channel_id, access_hash, receiver_id, include_keywords, exclude_keywords  FROM channels WHERE name = $1", channelName).Scan(&data.Name, &data.ChannelID, &data.AccessHash, &data.ReceiverID, &includeKeywordsJSON, &excludeKeywordsJSON)
	if err != nil {
		return data, err
	}

	if len(includeKeywordsJSON) > 0 {
		err = json.Unmarshal(includeKeywordsJSON, &data.IncludeKeywords)
		if err != nil {
			return data, err // Возвращаем ошибку при десериализации
		}
	}

	if len(excludeKeywordsJSON) > 0 {
		err = json.Unmarshal(excludeKeywordsJSON, &data.ExcludeKeywords)
		if err != nil {
			return data, err // Возвращаем ошибку при десериализации
		}
	}

	return data, nil
}

// Сохранение или обновление данных канала в БД
func SaveOrUpdateChannelData(db *sql.DB, channelData ChannelData) error {
	includeKeywordsJSON, err := json.Marshal(channelData.IncludeKeywords)
	if err != nil {
		return err // Возвращаем ошибку при сериализации ключевых слов в JSON
	}

	excludeKeywordsJSON, err := json.Marshal(channelData.ExcludeKeywords)
	if err != nil {
		return err // Возвращаем ошибку при сериализации ключевых слов в JSON
	}

	_, err = db.Exec("INSERT INTO channels (name, channel_id, access_hash, receiver_id, include_keywords, exclude_keywords) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (name) DO UPDATE SET channel_id = EXCLUDED.channel_id, access_hash = EXCLUDED.access_hash, receiver_id = EXCLUDED.receiver_id, include_keywords = EXCLUDED.include_keywords, exclude_keywords = EXCLUDED.exclude_keywords", channelData.Name, channelData.ChannelID, channelData.AccessHash, channelData.ReceiverID, includeKeywordsJSON, excludeKeywordsJSON)
	return err
}
