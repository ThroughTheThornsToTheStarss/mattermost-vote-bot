package models

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"time"
)

type Vote struct {
	ID        string         `json:"id"`
	CreatorID string         `json:"creator_id"`
	ChannelID string         `json:"channel_id"`
	Question  string         `json:"question"`
	Options   []string       `json:"options"`
	Votes     map[string]int `json:"votes"` // user_id -> option_index
	CreatedAt time.Time      `json:"created_at"`
	IsActive  bool           `json:"is_active"`
}

// NewVote создает новое голосование с проверкой входных данных
func NewVote(creatorID, channelID, question string, options []string) (*Vote, error) {
	if strings.TrimSpace(question) == "" {
		return nil, ErrEmptyQuestion
	}
	if len(options) < 2 {
		return nil, ErrNotEnoughOptions
	}

	return &Vote{
		ID:        generateID(),
		CreatorID: creatorID,
		ChannelID: channelID,
		Question:  question,
		Options:   options,
		Votes:     make(map[string]int),
		CreatedAt: time.Now().UTC(),
		IsActive:  true,
	}, nil
}

// CastVote добавляет или обновляет голос пользователя
func (v *Vote) CastVote(userID string, optionIndex int) error {
	if !v.IsActive {
		return ErrInactiveVote
	}
	if optionIndex < 0 || optionIndex >= len(v.Options) {
		return ErrInvalidOption
	}
	v.Votes[userID] = optionIndex
	return nil
}

// Results возвращает количество голосов за каждый вариант
func (v *Vote) Results() map[string]int {
	results := make(map[string]int)
	for _, index := range v.Votes {
		if index >= 0 && index < len(v.Options) {
			option := v.Options[index]
			results[option]++
		}
	}
	return results
}

// Завершение голосования
func (v *Vote) Close() {
	v.IsActive = false
}

// Генерация безопасного уникального ID
func generateID() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		panic("cannot generate ID") // в реальном коде лучше возвращать ошибку
	}
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}

// Ошибки
var (
	ErrEmptyQuestion    = errors.New("question cannot be empty")
	ErrNotEnoughOptions = errors.New("at least 2 options required")
	ErrInactiveVote     = errors.New("vote is not active")
	ErrInvalidOption    = errors.New("invalid option index")
)
