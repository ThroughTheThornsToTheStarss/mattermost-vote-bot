package db

import (
	"errors"
	"fmt"
	"time"

	"github.com/ThroughTheThornsToTheStarss/mattermost-vote-bot/internal/models"
	"github.com/tarantool/go-tarantool"
)

type DB struct {
	conn *tarantool.Connection
}

func New(addr string) (*DB, error) {
	conn, err := tarantool.Connect(addr, tarantool.Opts{
		Timeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Tarantool: %w", err)
	}
	return &DB{conn: conn}, nil
}

func (db *DB) GetVote(id string) (*models.Vote, error) {
	// Правильный вызов Select для последних версий go-tarantool
	resp, err := db.conn.Select("votes", "primary", 0, 1, 0, []interface{}{id})
	if err != nil {
		return nil, fmt.Errorf("tarantool select error: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, errors.New("vote not found")
	}

	data, ok := resp.Data[0].([]interface{})
	if !ok || len(data) < 8 {
		return nil, errors.New("invalid vote data format")
	}

	options := convertToStringSlice(data[4].([]interface{}))
	votes := convertToVotesMap(data[5].(map[interface{}]interface{}))

	return &models.Vote{
		ID:        data[0].(string),
		CreatorID: data[1].(string),
		ChannelID: data[2].(string),
		Question:  data[3].(string),
		Options:   options,
		Votes:     votes,
		CreatedAt: time.Unix(data[6].(int64), 0),
		IsActive:  data[7].(bool),
	}, nil
}

func (db *DB) SaveVote(vote *models.Vote) error {
	_, err := db.conn.Insert("votes", serializeVote(vote))
	if err != nil {
		return fmt.Errorf("failed to save vote: %w", err)
	}
	return nil
}

func (db *DB) UpdateVote(vote *models.Vote) error {
	_, err := db.conn.Replace("votes", serializeVote(vote))
	if err != nil {
		return fmt.Errorf("failed to update vote: %w", err)
	}
	return nil
}

func (db *DB) DeleteVote(id string) error {
	_, err := db.conn.Delete("votes", "primary", []interface{}{id})
	if err != nil {
		return fmt.Errorf("failed to delete vote: %w", err)
	}
	return nil
}

// ====== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ======

func serializeVote(v *models.Vote) []interface{} {
	return []interface{}{
		v.ID,
		v.CreatorID,
		v.ChannelID,
		v.Question,
		v.Options,
		v.Votes,
		v.CreatedAt.Unix(),
		v.IsActive,
	}
}

// convertToStringSlice преобразует []interface{} в []string
func convertToStringSlice(input []interface{}) []string {
	result := make([]string, len(input))
	for i, v := range input {
		result[i] = fmt.Sprint(v) // Безопасное преобразование любого типа в string
	}
	return result
}

// convertToVotesMap преобразует map[interface{}]interface{} в map[string]int
func convertToVotesMap(input map[interface{}]interface{}) map[string]int {
	result := make(map[string]int)
	for k, v := range input {
		key := fmt.Sprint(k)
		switch val := v.(type) {
		case int:
			result[key] = val
		case float64:
			result[key] = int(val)
		default:
			result[key] = 0
		}
	}
	return result
}

func (db *DB) Vote(userID, voteID string, optionIndex int) error {
	vote, err := db.GetVote(voteID)
	if err != nil {
		return err
	}

	err = vote.CastVote(userID, optionIndex)
	if err != nil {
		return err
	}

	return db.UpdateVote(vote)
}

// EndVote завершает голосование с проверкой создателя
func (db *DB) EndVote(userID, voteID string) error {
	// 1. Получаем голосование
	vote, err := db.GetVote(voteID)
	if err != nil {
		return fmt.Errorf("failed to get vote: %w", err)
	}

	// 2. Проверяем права
	if vote.CreatorID != userID {
		return errors.New("only vote creator can end it")
	}

	// 3. Обновляем статус
	vote.IsActive = false
	return db.UpdateVote(vote)
}
