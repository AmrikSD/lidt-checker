package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-lambda-go/lambda"
)

type Author struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	IconURL string `json:"icon_url"`
}

type Embed struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Colour int    `json:"color"`
	Author Author `json:"author"`
}

type DiscordMessage struct {
	Content   string  `json:"content,omitempty"`
	Embeds    []Embed `json:"embeds"`
	Username  string  `json:"username"`
	AvatarURL string  `json:"avatar_url"`
}

func handler() {
	status := getJapanStatus()
	updateDiscord(status)
}

func main() {
	lambda.Start(handler)
}

func getJapanStatus() string {

	response, err := http.Get("https://www.japan-guide.com/news/alerts.html")
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Error loading HTTP reponse body. ", err)
	}

	status := document.Find("div.post_list__content:nth-child(3) > p:nth-child(2)").First().Text()
	lastUpdated := document.Find(".post_list__date").First().Text()
	return fmt.Sprintf("%s\n%s", status, lastUpdated)
}

func checkEnvExists(str string) {
	_, exists := os.LookupEnv(str)
	if !exists {
		log.Fatal("Environment variable not set: ", str)
		os.Exit(1)
	}
}

func updateDiscord(status string) {

	webhookURL, _ := os.LookupEnv("WEBHOOK_URL")
	messageLink, _ := os.LookupEnv("MESSAGE_LINK")
	authorName, _ := os.LookupEnv("AUTHOR_NAME")
	authorURL, _ := os.LookupEnv("AUTHOR_URL")
	authorIcon, _ := os.LookupEnv("AUTHOR_ICON")
	messageColourStr := os.Getenv("MESSAGE_COLOUR")
	username, _ := os.LookupEnv("WEBHOOK_USERNAME")
	avatarURL, _ := os.LookupEnv("AVATAR_URL")

	messageColour, err := strconv.Atoi(messageColourStr)
	if err != nil {
		panic(err)
	}

	envVars := [8]string{
		"WEBHOOK_URL",
		"MESSAGE_LINK",
		"AUTHOR_NAME",
		"AUTHOR_URL",
		"AUTHOR_ICON",
		"MESSAGE_COLOUR",
		"WEBHOOK_USERNAME",
		"AVATAR_URL",
	}

	for _, str := range envVars {
		checkEnvExists(str)
	}

	msg := DiscordMessage{
		Embeds: []Embed{
			{
				Title:  status,
				URL:    messageLink,
				Colour: messageColour,
				Author: Author{
					Name:    authorName,
					URL:     authorURL,
					IconURL: authorIcon,
				},
			},
		},
		Username:  username,
		AvatarURL: avatarURL,
	}

	bytesRepresentation, err := json.Marshal(msg)
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

}
