package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/ThroughTheThornsToTheStarss/mattermost-vote-bot/internal/db"
	"github.com/ThroughTheThornsToTheStarss/mattermost-vote-bot/internal/models"
)

type Bot struct {
	db *db.DB
	mm *MattermostClient
}

func New(db *db.DB, mm *MattermostClient) *Bot {
	return &Bot{
		db: db,
		mm: mm,
	}
}

func (b *Bot) HandleCommand(userID, channelID, command string) {
	args := parseCommand(command)
	if len(args) == 0 {
		return
	}

	switch args[0] {
	case "create":
		b.handleCreate(userID, channelID, args[1:])
	case "end":
		b.handleEnd(userID, channelID, args[1:])
	case "results":
		b.handleResults(channelID, args[1:])
	case "delete":
		b.handleDelete(userID, channelID, args[1:])
	default:
		b.handleVote(userID, channelID, args)
	}

}

func (b *Bot) handleCreate(userID, channelID string, args []string) {
	if len(args) < 2 {
		b.mm.SendMessage(channelID, "Usage: `/vote create \"Question\" \"Option1\" \"Option2\"...`")
		return
	}

	vote, err := models.NewVote(userID, channelID, args[0], args[1:])
	if err != nil {
		b.mm.SendMessage(channelID, "Error: "+err.Error())
		return
	}

	if err := b.db.SaveVote(vote); err != nil {
		log.Printf("CreateVote error: %v", err)
		b.mm.SendMessage(channelID, "Failed to create vote")
		return
	}

	msg := fmt.Sprintf("Vote created (ID: `%s`)\nQuestion: %s\nOptions:\n%s",
		vote.ID, vote.Question, formatOptions(vote.Options))
	b.mm.SendMessage(channelID, msg)
}

func (b *Bot) handleEnd(userID, channelID string, args []string) {
	if len(args) < 1 {
		b.mm.SendMessage(channelID, "Usage: `/vote end [voteID]`")
		return
	}

	if err := b.db.EndVote(userID, args[0]); err != nil {
		b.mm.SendMessage(channelID, "Error: "+err.Error())
		return
	}

	b.mm.SendMessage(channelID, "Vote closed successfully")
}

func (b *Bot) handleVote(userID, channelID string, args []string) {
	if len(args) != 2 {
		b.mm.SendMessage(channelID, "Usage: `/vote [voteID] [optionNumber]`")
		return
	}

	optionIdx, err := strconv.Atoi(args[1])
	if err != nil || optionIdx < 1 {
		b.mm.SendMessage(channelID, "Option must be a positive number")
		return
	}

	if err := b.db.Vote(userID, args[0], optionIdx-1); err != nil {
		b.mm.SendMessage(channelID, "Error: "+err.Error())
		return
	}

	b.mm.SendMessage(channelID, "Your vote has been recorded")
}

func parseCommand(cmd string) []string {
	cmd = strings.TrimSpace(strings.TrimPrefix(cmd, "/vote"))
	return strings.Fields(cmd)
}

func formatOptions(options []string) string {
	var sb strings.Builder
	for i, opt := range options {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, opt))
	}
	return sb.String()
}

func (b *Bot) handleResults(channelID string, args []string) {
	if len(args) < 1 {
		b.mm.SendMessage(channelID, "Usage: `/vote results [voteID]`")
		return
	}

	vote, err := b.db.GetVote(args[0])
	if err != nil {
		b.mm.SendMessage(channelID, "Error: "+err.Error())
		return
	}

	results := vote.Results()
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Results for '%s':\n", vote.Question))
	for option, count := range results {
		sb.WriteString(fmt.Sprintf("- %s: %d votes\n", option, count))
	}
	b.mm.SendMessage(channelID, sb.String())
}
func (b *Bot) handleDelete(userID, channelID string, args []string) {
	if len(args) < 1 {
		b.mm.SendMessage(channelID, "Usage: `/vote delete [voteID]`")
		return
	}

	vote, err := b.db.GetVote(args[0])
	if err != nil {
		b.mm.SendMessage(channelID, "Error: "+err.Error())
		return
	}

	if vote.CreatorID != userID {
		b.mm.SendMessage(channelID, "Only the creator can delete this vote")
		return
	}

	if err := b.db.DeleteVote(args[0]); err != nil {
		b.mm.SendMessage(channelID, "Error: "+err.Error())
		return
	}

	b.mm.SendMessage(channelID, "Vote deleted successfully")
}
