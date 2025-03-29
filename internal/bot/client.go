package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
)

var (
	ErrMattermostConnection = errors.New("failed to connect to Mattermost")
)

type MattermostClient struct {
	client *model.Client4
	ws     *model.WebSocketClient
}

func NewMattermostClient() (*MattermostClient, error) {
	token := os.Getenv("MM_TOKEN")
	if token == "" {
		return nil, errors.New("MM_TOKEN environment variable not set")
	}

	client := model.NewAPIv4Client(os.Getenv("MM_URL"))
	if client == nil {
		return nil, ErrMattermostConnection
	}
	client.SetToken(token)

	_, _, err := client.GetMe("")
	if err != nil {
		return nil, fmt.Errorf("GetMe error: %v", err)
	}

	wsURL := strings.Replace(os.Getenv("MM_URL"), "http", "ws", 1)
	ws, err := model.NewWebSocketClient4(wsURL, token)
	if err != nil {
		return nil, err
	}

	return &MattermostClient{
		client: client,
		ws:     ws,
	}, nil
}

func (mc *MattermostClient) Listen(bot *Bot) {
	mc.ws.Listen()
	defer mc.ws.Close()

	for event := range mc.ws.EventChannel {
		log.Printf("event : %v", event)
		if event.EventType() == model.WebsocketEventPosted {
			mc.handlePostEvent(event, bot)
		}
	}
}

func (mc *MattermostClient) handlePostEvent(event *model.WebSocketEvent, bot *Bot) {
	postData, ok := event.GetData()["post"].(string)
	if !ok {
		return
	}

	var post *model.Post
	if err := json.Unmarshal([]byte(postData), &post); err != nil {
		return
	}

	if post != nil && strings.HasPrefix(post.Message, "/vote") {
		bot.HandleCommand(post.UserId, post.ChannelId, post.Message)
	}
}

func (mc *MattermostClient) SendMessage(channelID, message string) error {
	post := &model.Post{
		ChannelId: channelID,
		Message:   message,
	}
	_, _, err := mc.client.CreatePost(post)
	if err != nil {
		return fmt.Errorf("CreatePost error: %v", err)
	}
	return nil
}
