package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/krognol/go-wolfram"
	"github.com/shomali11/slacker"
	"github.com/tidwall/gjson"
	witai "github.com/wit-ai/wit-go"
)

var wolframClient *wolfram.Client

func PrintCommandEvents(analyticsChannel <-chan *slacker.CommandEvent) {
	for event := range analyticsChannel {
		fmt.Println("CommandEvent")
		fmt.Println(event.Timestamp)
		fmt.Println(event.Command)
		fmt.Println(event.Parameters)
		fmt.Println(event.Event)
		fmt.Println()
	}
}
func main() {
	godotenv.Load(".env")
	bot := slacker.NewClient(os.Getenv("SLACK_BOT_TOKEN"), os.Getenv("SLACK_APP_TOKEN"))
	client := witai.NewClient(os.Getenv("WIT_AI_TOKEN"))
	wolframClient = &wolfram.Client{AppID: os.Getenv("WOLFRAM_APP_ID")}
	go PrintCommandEvents(bot.CommandEvents())

	bot.Command("<message>", &slacker.CommandDefinition{
		Description: "send any question",
		Examples:    []string{"Who is the president of Argentina?"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			query := request.Param("message")
			//fmt.Println(query)
			msg, err := client.Parse(&witai.MessageRequest{
				Query: query,
			})
			if err != nil {
				fmt.Println(err)
				response.Reply(err.Error())
				return
			}
			fmt.Println(msg)
			data, err := json.MarshalIndent(msg, "", "    ")
			if err != nil {
				fmt.Println(err)
				response.Reply(err.Error())
				return
			}

			rough := string(data[:])
			fmt.Println(rough)
			value := gjson.Get(rough, "entities.wolfram_search_query.0.value")
			answer := value.String()
			resp, err := wolframClient.GetSpokentAnswerQuery(answer, wolfram.Metric, 1000)

			if err != nil {
				fmt.Println(err)
				response.Reply(err.Error())
				return
			}
			response.Reply(resp)
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := bot.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}

}
