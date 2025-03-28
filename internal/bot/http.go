package bot

import (
	"fmt"
	"net/http"
)

func (b *Bot) StartHTTPServer() {
	http.HandleFunc("/vote", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		userID := r.PostFormValue("user_id")
		channelID := r.PostFormValue("channel_id")
		text := r.PostFormValue("text")

		go b.HandleCommand(userID, channelID, "vote "+text)

		fmt.Fprintf(w, "Обрабатываю команду: %s", text)
	})

	go http.ListenAndServe(":5000", nil)
}
