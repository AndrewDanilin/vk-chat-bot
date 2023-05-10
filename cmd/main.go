package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"VKInternshipChatBot/internal/adapters/reminderrepo"
	"VKInternshipChatBot/internal/app"
	"VKInternshipChatBot/internal/ports"
)

const (
	vk_api = "https://api.vk.com/method/"
	port   = ":18080"
)

type ChatBotConfig struct {
	AccessToken  string `json:"access_token"`
	GroupID      string `json:"group_id"`
	VKApiVersion string `json:"vk_api_version"`
}

func main() {
	configFile := flag.String("config", "", "bot configuration file")

	flag.Parse()

	if configFile == nil || *configFile == "" {
		log.Fatal("config file not specified")
	}

	bytes, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("read config file error: %s\n", err)
	}

	cfg := ChatBotConfig{}

	if err := json.Unmarshal(bytes, &cfg); err != nil {
		log.Fatalf("broken config file: %s\n", err)
	}

	log.Println("config file successfully parsed")

	client := ports.NewClient(vk_api, cfg.AccessToken, app.NewApp(reminderrepo.New()), cfg.GroupID, cfg.VKApiVersion)
	if err != nil {
		log.Fatalf("failed to create LongPollServer: %s\n", err)
	}

	if err := client.Run(); err != nil {
		log.Fatalln(err)
	}
}
