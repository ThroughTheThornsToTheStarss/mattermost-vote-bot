package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ThroughTheThornsToTheStarss/mattermost-vote-bot/internal/bot"
	"github.com/ThroughTheThornsToTheStarss/mattermost-vote-bot/internal/db"
)

func main() {
	// Подключаем Tarantool
	tarantoolDB, err := db.New("tarantool:3301")
	if err != nil {
		log.Fatal("Ошибка подключения к Tarantool:", err)
	}

	// Подключаемся к Mattermost
	mmClient, err := bot.NewMattermostClient()
	if err != nil {
		log.Fatal("Ошибка подключения к Mattermost:", err)
	}

	// Создаем бота с зависимостями
	voteBot := bot.New(tarantoolDB, mmClient)
	// запуск сервера

	// ✅ Запускаем HTTP-сервер для Slash-команд
	voteBot.StartHTTPServer()
	// Запускаем слушателя команд в отдельной горутине

	go mmClient.Listen(voteBot)

	// Ожидаем сигналов завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	log.Println("Бот запущен. Для остановки нажмите Ctrl+C.")
	<-sigChan
	log.Println("Завершение работы бота...")
}
