package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-lambda-go/lambda"
)

type Author struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	IconURL string `json:"icon_url"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type Image struct {
	URL string `json:"url"`
}

type Footer struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url"`
}

type Embed struct {
	Author      Author  `json:"author"`
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Description string  `json:"description"`
	Colour      int     `json:"color"`
	Fields      []Field `json:"fields,omitempty"`
	Thumbnail   Image   `json:"thumbnail,omitempty"`
	Image       Image   `json:"image,omitempty"`
	Footer      Footer  `json:"footer,omitempty"`
}

type DiscordMessage struct {
	Username  string  `json:"username"`
	AvatarURL string  `json:"avatar_url"`
	Content   string  `json:"content,omitempty"`
	Embeds    []Embed `json:"embeds"`
}

func handler() {
	availableTests := getLidtTests("Tolworth (London)")
	if len(availableTests) > 0 {
		updateDiscord(availableTests)
	}
}

func main() {
	environment := os.Getenv("APP_ENV")
	if environment == "DEV" {
		handler()
	} else if environment == "PROD" {
		lambda.Start(handler)
	} else {
		log.Fatal("Environment not correct")
	}
}

func getLidtTests(centre string) []Field {

	response, err := http.Get("https://lidt.co.uk/fast-track-booking")

	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Fatal("Error loading HTTP reponse body.", err)
	}

	var testInWantedLocation []Field

	testOptions := document.Find("#testCentres > option")
	testOptions.Each(func(_ int, testOption *goquery.Selection) {
		location, exists := testOption.Attr("data-location")
		time := testOption.AttrOr("data-start", "unknown")

		if exists && location == centre {
			testInWantedLocation = append(testInWantedLocation, Field{
				Name:  location,
				Value: time,
			})
		} else if strings.Contains(testOption.Text(), "Tolworth") {
			testInWantedLocation = append(testInWantedLocation, Field{
				Name:  "Tolworth (maybe)",
				Value: "Unknown",
			})
		}
	})

	return testInWantedLocation
}

func checkEnvExists(str string) {
	_, exists := os.LookupEnv(str)
	if !exists {
		log.Fatalf("Environment variable not set: %s", str)
		os.Exit(1)
	}
}

func updateDiscord(fields []Field) {

	checkEnvExists("WEBHOOK_URL")
	webhookURL, _ := os.LookupEnv("WEBHOOK_URL")

	msg := DiscordMessage{
		Username:  "Asbo - Webhook",
		AvatarURL: "https://cdn.discordapp.com/avatars/122650490289913856/0da6bbf0c4c1c7006b578dd189a4c244",
		Content:   "<@&681217602432794627>",
		Embeds: []Embed{
			{
				Author: Author{
					Name:    "AmrikSD",
					URL:     "https://github.com/lidt-checker",
					IconURL: "https://avatars.githubusercontent.com/u/17920436",
				},
				Title:  "Wanted Tests Available!",
				URL:    "https://lidt.co.uk/fast-track-booking",
				Colour: 12345678,
				Fields: fields, //Passed into the method
				Footer: Footer{
					Text:    "https://github.com/AmrikSD/lidt-checker",
					IconURL: "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png",
				},
			},
		},
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
