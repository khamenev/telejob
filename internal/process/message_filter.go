package process

import "strings"

type Message struct {
	ID         int64
	Text       string
	ChannelId  int64
	ReceiverId int64
}

type MessageFilter struct {
	IncludeKeywords []string
	ExcludeKeywords []string
}

func NewMessageFilter(includeKeywords []string, excludeKeywords []string) *MessageFilter {
	// Приводим ключевые слова к нижнему регистру при создании фильтра
	lowerCaseIncludeKeywords := make([]string, len(includeKeywords))
	for i, keyword := range includeKeywords {
		lowerCaseIncludeKeywords[i] = strings.ToLower(keyword)
	}

	lowerCaseExcludeKeywords := make([]string, len(excludeKeywords))
	for i, keyword := range excludeKeywords {
		lowerCaseExcludeKeywords[i] = strings.ToLower(keyword)
	}

	return &MessageFilter{
		IncludeKeywords: lowerCaseIncludeKeywords,
		ExcludeKeywords: lowerCaseExcludeKeywords,
	}
}

func (mf *MessageFilter) Filter(messages []Message) []Message {
	var filtered []Message
	for _, msg := range messages {
		if mf.containsIncludeKeyword(msg.Text) && !mf.containsExcludeKeyword(msg.Text) {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

func (mf *MessageFilter) containsIncludeKeyword(text string) bool {
	lowerCaseText := strings.ToLower(text) // Приводим текст сообщения к нижнему регистру
	for _, keyword := range mf.IncludeKeywords {
		if strings.Contains(lowerCaseText, keyword) { // Используем приведенный к нижнему регистру текст для сравнения
			return true
		}
	}
	return false
}

func (mf *MessageFilter) containsExcludeKeyword(text string) bool {
	lowerCaseText := strings.ToLower(text)
	for _, keyword := range mf.ExcludeKeywords {
		if strings.Contains(lowerCaseText, keyword) {
			return true
		}
	}
	return false
}
