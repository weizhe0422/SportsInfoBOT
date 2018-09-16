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
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/weizhe0422/SportsInfoBOT/cpblschedule"
)

var bot *linebot.Client

func main() {
	var err error
	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
	log.Println("Bot:", bot, " err:", err)
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	choiceList := []string{"1. 目前賽事比分", "2. 本月賽事"}

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
		listAllContents(bot, event.ReplyToken, choiceList)

		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				switch message.Text {
				case "1. 目前賽事比分":
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("尚未完成")).Do(); err != nil {
						log.Println(err)
					}
				case "2. 本月賽事":
					result := make([]cpblschedule.Match, 0)
					result, _ = cpblschedule.ParseCPBLSchedule(2018, 9)
					var resultString string

					for _, item := range result {
						tmpItem := item.Date + "/" + item.Location + "/" + item.Home + "(" + item.Home_score + ")" + "/" + item.Away + "(" + item.Away_score + ")"
						resultString = resultString + tmpItem + "||"
					}

					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(resultString)).Do(); err != nil {
						log.Println(err)
					}
				}

				//if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.ID+":"+message.Text+" OK!")).Do(); err != nil {
				//	cpblschedule.ParseCPBLSchedule(2018, 9)
				//	log.Print(err)
				//}
			}
		}
	}
}

func listAllContents(bot *linebot.Client, replyToken string, intentList []string) {
	var sliceTemplateAction []linebot.TemplateAction

	for _, v := range intentList {
		sliceTemplateAction = append(sliceTemplateAction, linebot.NewPostbackAction(v, v, v, ""))
	}

	template := linebot.NewButtonsTemplate("", "你想選擇什麼？", "", sliceTemplateAction...)
	if _, err := bot.ReplyMessage(replyToken, linebot.NewTemplateMessage("你想選擇什麼？", template)).Do(); err != nil {
		log.Println(err)
	}
}
