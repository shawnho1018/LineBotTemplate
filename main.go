// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
)

var bot *linebot.Client
var base_url string
var webhooks map[string]map[string]interface{}

func main() {
	var err error
	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
	base_url = os.Getenv("APP_BASE_URL")
	log.Println("base_url: ", base_url)
	log.Println("Bot:", bot, " err:", err)
	// load webhook list
	byteVal, err := ioutil.ReadFile(webhooks.json)
	if err != nil {
		log.Fatal(err)
		return
	}
	if err := json.Unmarshal(byteVal, &webhooks); err != nil {
		log.Fatal(err)
		return
	}
	for k, v := range webhooks {
		fmt.Printf("%s -> %s\n", k, v)
		for k1, v1 := range v {
			fmt.Printf("%s -> %s\n", k1, v1)
		}
	}
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}
func callbuild(url string, token string, ref string) {
	var r http.Request
	r.ParseForm()
	r.Form.Add("token", token)
	r.Form.Add("ref", ref)
	bodystr := strings.TrimSpace(r.Form.Encode())
	request, err := http.NewRequest("POST", url, strings.NewReader(bodystr))
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	var resp *http.Response
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	byts, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(string(byts))
}
func callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		log.Printf("Got event %v", event)
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if err := handleText(message, event.ReplyToken, event.Source); err != nil {
					log.Print(err)
				}
			}
		}
	}
}
func replyText(replyToken, text string) error {
	if _, err := bot.ReplyMessage(
		replyToken,
		linebot.NewTextMessage(text),
	).Do(); err != nil {
		return err
	}
	return nil
}
func handleText(message *linebot.TextMessage, replyToken string, source *linebot.EventSource) error {
	switch message.Text {
	case "Profile":
		if source.UserID != "" {
			profile, err := bot.GetProfile(source.UserID).Do()
			if err != nil {
				return replyText(replyToken, err.Error())
			}
			if _, err := bot.ReplyMessage(
				replyToken,
				linebot.NewTextMessage("Display name: "+profile.DisplayName),
				linebot.NewTextMessage("Status message: "+profile.StatusMessage),
			).Do(); err != nil {
				return err
			}
		} else {
			return replyText(replyToken, "Bot can't use profile API without user ID")
		}
	case "Build":
		imageURL := "https://miro.medium.com/max/1500/1*pgvyLv6PdRGC54BpV3POcQ.png"
		//log.Println("Tanzu Image Path:", imageURL)
		template := linebot.NewButtonsTemplate(
			imageURL, "Build Sample", "Hello! What would you like to build today?",
			linebot.NewURIAction("Go to line.me", "https://line.me"),
			linebot.NewPostbackAction("Say hello1", "hello こんにちは", "", "hello こんにちは"),
			linebot.NewPostbackAction("言 hello2", "hello こんにちは", "hello こんにちは", ""),
			linebot.NewMessageAction("Say message", "Rice=米"),
		)
		if _, err := bot.ReplyMessage(
			replyToken,
			linebot.NewTemplateMessage("Buttons alt text", template),
		).Do(); err != nil {
			return err
		}
	case "Build1":
		template := linebot.NewConfirmTemplate(
			"Do it?",
			linebot.NewMessageAction("Yes", "Yes!"),
			linebot.NewMessageAction("No", "No!"),
		)
		if _, err := bot.ReplyMessage(
			replyToken,
			linebot.NewTemplateMessage("Confirm alt text", template),
		).Do(); err != nil {
			return err
		}
	case "Yes!":
		callbuild("https://gitlab.com/api/v4/projects/16654842/trigger/pipeline", "3500c5b9724537c6bc182eaa5642bc", "master")
	default:
		log.Printf("Echo message to %s: %s", replyToken, message.Text)
		if _, err := bot.ReplyMessage(
			replyToken,
			linebot.NewTextMessage(message.Text),
		).Do(); err != nil {
			return err
		}
	}
	return nil
}
