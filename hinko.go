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
	"strconv"
	"strings"
	"syscall"

	"github.com/tadej/hinko/ascii"
	"github.com/tadej/hinko/model"
	"github.com/tadej/hinko/slack"
)

type processCommand func([]string, slack.MessageInfo) string

var emojiCommandNotFound = "shrug"                  // 🤷‍♀️
var emojiParametersWrong = "heavy_multiplication_x" // ✖️
var emojiCommandError = "bug"                       // 🐛
var emojiCommandOK = "heavy_check_mark"             // ✔️
var emojiCommandWarning = "grey_question"           // ❔

// this DB key contains a list of space-delimited team names
var teamNamesGroup = "teamnames"

// this DB key contains a list of space-delimited pair names
var pairNamesGroup = "pairnames"

// accepted text commands with their corresponding processor functions
var acceptedCommands = map[string]processCommand{
	"put":         processCommandPut,
	"get":         processCommandGet,
	"shark":       processCommandShark,
	"animate":     processCommandAnimate,
	"help":        processCommandHelp,
	"randompairs": processCommandRandomPairs,
	"randomteams": processCommandRandomTeams,
	"group":       processCommandGroup,
	"ascii":       processCommandASCII,
}

func main() {
	token := os.Getenv("SLACK_TOKEN")
	slack.Init(token)
	model.OpenDatabase("/tmp/model.lvl")
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
	var returnMessage string
	var prefix string

	// mention, tag the user in the channel
	if !msg.IM {
		prefix = "<@" + msg.UserID + ">\n"
	}

	parts := strings.Split(message, " ")
	if len(parts) > 0 {
		fn := acceptedCommands[parts[0]]
		if fn != nil {
			returnMessage = fn(parts, msg)
		}
		return prefix + returnMessage
	}

	// command not supported
	react(msg, emojiCommandNotFound)

	return ""
}

func react(msg slack.MessageInfo, reaction string) {
	slack.AddReaction(msg.Username, msg.Channel, msg.Timestamp, reaction)
}

func processCommandHelp(parts []string, msg slack.MessageInfo) string {
	infoMessage :=
		`Try the following commands:` + "\n" +
			"`help`\n" +
			"`put key value`\n" +
			"`get key`\n" +
			"`group groupname list`\n" +
			"`group groupname create @user1 @user2 @user3 ...`\n" +
			"`group groupname add @user1 @user2 @user3 ...`\n" +
			"`group groupname remove @user1 @user2 @user3 ...`\n" +
			"`randompairs @user1 @user2 @user3 ...`\n" +
			"`randompairs group`\n" +
			"`randomteams teamsize @user1 @user2 @user3 ...`\n" +
			"`randomteams teamsize group`\n" +
			"`ascii`\n" +
			"`shark`\n" +
			"`animate`\n\n" +
			" reserved groups: _pairnames_, _teamnames_\n\n" +
			"More info:\nhttps://github.com/tadej/hinko"

	return infoMessage
}

func processCommandASCII(parts []string, msg slack.MessageInfo) string {
	return ascii.ImageToASCII(parts[1])
}

func processCommandShark(parts []string, msg slack.MessageInfo) string {
	go ascii.DoSharkAnimation(30, 2, 300,
		func(txt string) (string, string) {
			channel, timestamp, _ := slack.PostMessage(msg.Channel, txt)
			return channel, timestamp
		}, func(channel string, timestamp string, newTxt string) {
			_ = slack.UpdateMessage(channel, timestamp, newTxt)
		})

	return ""
}

func processCommandAnimate(parts []string, msg slack.MessageInfo) string {
	go ascii.DoFrameAnimation(30, 300,
		func(txt string) (string, string) {
			channel, timestamp, _ := slack.PostMessage(msg.Channel, txt)
			return channel, timestamp
		}, func(channel string, timestamp string, newTxt string) {
			_ = slack.UpdateMessage(channel, timestamp, newTxt)
		})

	return ""
}

func processCommandGroup(parts []string, msg slack.MessageInfo) string {
	var returnMessage string

	if len(parts) < 3 {
		react(msg, emojiParametersWrong)
		return ""
	}

	groupName := parts[1]

	switch parts[2] {
	case "list":

	case "create":
		err := model.SetGroup(groupName, parts[3:])
		if err != nil {
			react(msg, emojiParametersWrong)
			return ""
		}
	case "add":
		err := model.AddToGroup(groupName, parts[3:])
		if err != nil {
			react(msg, emojiParametersWrong)
			return ""
		}

	case "remove":
		err := model.RemoveFromGroup(groupName, parts[3:])
		if err != nil {
			react(msg, emojiParametersWrong)
			return ""
		}
	default:
		react(msg, emojiParametersWrong)
		return ""

	}

	react(msg, emojiCommandOK)

	groupTest, err := model.GetGroup(parts[1])
	if err != nil {
		react(msg, emojiParametersWrong)
		return ""
	}
	returnMessage = "`" + parts[1] + "` members: " + strings.Join(groupTest, " ")

	return returnMessage
}

func processCommandRandomPairs(parts []string, msg slack.MessageInfo) string {
	var returnMessage string

	if len(parts) < 2 {
		react(msg, emojiParametersWrong)
		return ""
	}

	var members []string
	var err error
	// only one parameter means group name
	if len(parts) == 2 {
		members, err = model.GetGroup(parts[1])
		if err != nil {
			react(msg, emojiParametersWrong)
			return ""
		}
	} else {
		members = parts[1:]
	}

	teamNames, _ := model.GetGroup(pairNamesGroup)

	returnMessage, err = model.GetRandomTeams(2, members, true, teamNames, false)
	if err != nil {
		react(msg, emojiParametersWrong)
		return ""
	}

	return returnMessage
}

func processCommandRandomTeams(parts []string, msg slack.MessageInfo) string {
	var returnMessage string

	if len(parts) < 3 {
		react(msg, emojiParametersWrong)
		return ""
	}

	teamSize, err := strconv.Atoi(parts[1])
	if err != nil {
		react(msg, emojiParametersWrong)
		return ""
	}

	var members []string

	// only one parameter means we treat it as a group name
	if len(parts) == 3 {
		members, err = model.GetGroup(parts[2])
		if err != nil {
			react(msg, emojiParametersWrong)
			return ""
		}
	} else {
		members = parts[2:]
	}

	teamNames, _ := model.GetGroup(teamNamesGroup)

	returnMessage, err = model.GetRandomTeams(teamSize, members, false, teamNames, true)
	if err != nil {
		react(msg, emojiParametersWrong)
		return ""
	}

	return returnMessage
}

func processCommandPut(parts []string, msg slack.MessageInfo) string {
	if len(parts) < 3 {
		react(msg, emojiParametersWrong)
		return ""
	}

	err := model.SetDBValue(parts[1], parts[2])
	if err == nil {
		react(msg, emojiCommandOK)
	} else {
		react(msg, emojiCommandError)
	}

	return ""
}

func processCommandGet(parts []string, msg slack.MessageInfo) string {
	var returnMessage string

	if len(parts) < 2 {
		react(msg, emojiParametersWrong)
		return ""
	}

	data, err := model.GetDBValue(parts[1])
	if err == nil {
		returnMessage = data
	} else {
		react(msg, emojiCommandWarning)
	}
	return returnMessage
}
