package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/maxwww/family_bot/bot"
	"github.com/maxwww/family_bot/postgres"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	KievLocation = "Europe/Kiev"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}
	token := os.Getenv("TOKEN")
	postgresURL := os.Getenv("POSTGRESQL_URL")

	subscribersIdsString := os.Getenv("SUBSCRIBERS_IDS")
	subscribersIds := strings.Split(subscribersIdsString, ",")
	var subscribers []int64
	for _, v := range subscribersIds {
		id, _ := strconv.Atoi(v)
		subscribers = append(subscribers, int64(id))
	}
	if len(subscribers) == 0 {
		panic("it needs to specify SUBSCRIBERS_IDS")
	}

	location := os.Getenv("LOCATION")
	if location == "" {
		location = KievLocation
	}
	loc, err := time.LoadLocation(location)
	if err != nil {
		panic("it needs valid location")
	}

	botApi, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	db, err := postgres.Open(postgresURL)
	if err != nil {
		log.Fatalf("cannot open database: %v", err)
	}

	b := bot.NewBot(botApi, db, subscribers, loc)

	err = b.Start()
	if err != nil {
		log.Fatal(err)
	}
}
