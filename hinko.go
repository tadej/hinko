//MIT License

//Copyright(c) 2019 Tadej Gregorcic

//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:

//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.

//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.

package main

import (
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/tadej/hinko/commands"
	"github.com/tadej/hinko/model"
	"github.com/tadej/hinko/slack"
)

func main() {
	token := os.Getenv("SLACK_TOKEN")
	slack.Init(token)
	model.OpenDatabase(os.Getenv("DATABASE_PATH"))
	initInterruptHandler()

	fmt.Println("Starting Slack API loop")

	c := make(chan slack.MessageInfo)
	go slack.MessageLoop(c)

	for {
		message := <-c

		if message.OK {
			respond(message)
		} else {
			break
		}
	}

	defer model.CloseDatabase()
}

func initInterruptHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\rctrl+c pressed in terminal")
		model.CloseDatabase()
		os.Exit(0)
	}()
}

func respond(msg slack.MessageInfo) {
	var response string
	text := msg.Message

	text = strings.TrimPrefix(text, msg.Prefix)
	text = strings.TrimSpace(text)

	var mentionedBot = strings.HasPrefix(msg.Message, "<@"+msg.MyID+">")

	if msg.IM || mentionedBot {
		response = processMessage(text, msg)
		if response != "" {
			slack.SendMessage(msg.Channel, response)
		}
	}
}

func processMessage(message string, msg slack.MessageInfo) string {
	parts := strings.Split(message, " ")
	if len(parts) > 0 {
		fn := commands.[AcceptedCommandsstrings.ToLower(parts[0])]
		if fn != nil {
			return fn(parts, msg)
		}
	}

	// command not supported
	commands.React(msg, commands.EmojiCommandNotFound)

	return ""
}
