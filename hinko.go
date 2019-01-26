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

// Thanks to Spencer Smith for the tutorial: https://rsmitty.github.io/Slack-Bot/
// Thanks https://github.com/nlopes/slack

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nlopes/slack"
	"github.com/syndtr/goleveldb/leveldb"
)

func main() {

	token := os.Getenv("SLACK_TOKEN")
	api := slack.New(token)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

	db, err := leveldb.OpenFile("database.lvl", nil)
	if err != nil {
		fmt.Println(err)
		return
	}

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:

			case *slack.MessageEvent:
				info := rtm.GetInfo()

				user, err := api.GetUserInfo(ev.User)
				if err != nil {
					continue
				}

				var directMessage = isIMChannel(api, ev.Channel)

				prefix := fmt.Sprintf("<@%s> ", info.User.ID)

				if ev.User != info.User.ID {
					respond(db, info.User.ID, rtm, ev, prefix, user.ID, directMessage)
				}

			case *slack.RTMError:
				fmt.Printf("Error: %s\n\n", ev.Error())

			case *slack.InvalidAuthEvent:
				fmt.Printf("Invalid credentials")
				break Loop

			default:
			}
		}
	}
	defer db.Close()
}

func getDBValue(db *leveldb.DB, key string) (string, error) {
	data, err := db.Get([]byte(key), nil)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(data), nil
}

func setDBValue(db *leveldb.DB, key string, value string) error {
	err := db.Put([]byte(key), []byte(value), nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func isIMChannel(api *slack.Client, channel string) bool {
	var ret bool
	chans, err := api.GetIMChannels()
	if err != nil {
		fmt.Printf("%s\n", err)
		return false
	}
	for _, imchan := range chans {
		if imchan.ID == channel {
			ret = true
		}
	}
	return ret
}

func respond(db *leveldb.DB, botID string, rtm *slack.RTM, msg *slack.MessageEvent, prefix string, user string, directMessage bool) {
	var response string
	text := msg.Text

	text = strings.TrimPrefix(text, prefix)
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	var mentionedBot = strings.HasPrefix(msg.Text, "<@"+botID+">")

	if directMessage || mentionedBot {
		response = processMessage(text, user, directMessage, msg, rtm, db)
		if response != "" {
			rtm.SendMessage(rtm.NewOutgoingMessage(response, msg.Channel))
		}
	}
}

func getSharkString(pos int, len int, right bool) string {
	var ret string

	for i := 0; i < len; i++ {
		if pos == i {
			if right {
				ret += "\\"
			} else {
				ret += "|"
			}
		} else if pos == i+1 {
			if right {
				ret += "|"
			} else {
				ret += "/"
			}

		} else {
			ret += "_"
		}
	}

	return ret
}

func getAnimationFrame(i int) string {
	frames := [4]string{
		"╔════╤╤╤╤════╗\n" +
			"║    │││ \\   ║\n" +
			"║    │││  O  ║\n" +
			"║    OOO     ║",
		"╔════╤╤╤╤════╗\n" +
			"║    ││││    ║\n" +
			"║    ││││    ║\n" +
			"║    OOOO    ║",
		"╔════╤╤╤╤════╗\n" +
			"║   / │││    ║\n" +
			"║  O  │││    ║\n" +
			"║     OOO    ║",
		"╔════╤╤╤╤════╗\n" +
			"║    ││││    ║\n" +
			"║    ││││    ║\n" +
			"║    OOOO    ║"}

	return frames[i%len(frames)]
}

func sharkProc(channel string, rtm *slack.RTM, len int, maxTurns int) {
	var shark string
	var right bool

	right = false
	shark = getSharkString(0, len, right)

	retChan, retTimeStamp, err := rtm.PostMessage(channel, slack.MsgOptionText(shark, false), slack.MsgOptionAsUser(true))
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	for turns := 0; turns < maxTurns; turns++ {
		right = !right
		for i := 1; i < len-1; i++ {
			timer := time.NewTimer(100 * time.Millisecond)
			<-timer.C
			var newMsg string
			if right {
				newMsg = getSharkString(i, len, right)
			} else {
				newMsg = getSharkString(len-i, len, right)
			}
			_, _, _, err = rtm.UpdateMessage(retChan, retTimeStamp, slack.MsgOptionText(newMsg, false), slack.MsgOptionAsUser(true))
			if err != nil {
				fmt.Printf("%s\n", err)
			}
		}
	}
	newMsg := getSharkString(-1, len, right)
	_, _, _, err = rtm.UpdateMessage(retChan, retTimeStamp, slack.MsgOptionText(newMsg, false), slack.MsgOptionAsUser(true))
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
}

func animateProc(channel string, rtm *slack.RTM, len int) {
	var anim string
	anim = getAnimationFrame(0)
	retChan, retTimeStamp, err := rtm.PostMessage(channel, slack.MsgOptionText(anim, false), slack.MsgOptionAsUser(true))
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	for i := 1; i < len; i++ {
		timer := time.NewTimer(100 * time.Millisecond)
		<-timer.C
		var newMsg string
		newMsg = getAnimationFrame(i)
		_, _, _, err = rtm.UpdateMessage(retChan, retTimeStamp, slack.MsgOptionText(newMsg, false), slack.MsgOptionAsUser(true))
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
	}
}

func addReaction(msg *slack.MessageEvent, reaction string, rtm *slack.RTM) {
	var itemRef slack.ItemRef
	itemRef.Channel = msg.Channel
	itemRef.Timestamp = msg.Timestamp
	err := rtm.AddReaction(reaction, itemRef)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
}

func getInfoMessage() string {
	infoMessage :=
		`********************************
*Hinko the friendly Slack Bot*
********************************	
COMMANDS
help
put key value
get key
shark
animate

https://github.com/tadej/hinko
`
	return infoMessage
}

func processMessage(message string, userID string, directMessage bool, msg *slack.MessageEvent, rtm *slack.RTM, db *leveldb.DB) string {
	var returnMessage string
	var prefix string

	// mention, tag the user in the channel
	if !directMessage {
		prefix = "<@" + userID + "> "
	}

	parts := strings.Split(message, " ")

	acceptedCommands := map[string]bool{
		"put":     true,
		"get":     true,
		"shark":   true,
		"animate": true,
		"help":    true,
	}

	emojiCommandNotFound := "shrug"
	emojiParametersWrong := "heavy_multiplication_x"
	emojiCommandError := "bug"
	emojiCommandOK := "heavy_check_mark"
	emojiCommandWarning := "grey_question"

	infoMessage := getInfoMessage()

	if len(parts) < 1 || !acceptedCommands[parts[0]] {
		addReaction(msg, emojiCommandNotFound, rtm)
	}

	switch parts[0] {
	case "help":
		return infoMessage
	case "put":
		if len(parts) < 3 {
			addReaction(msg, emojiParametersWrong, rtm)
			return ""
		} else {
			err := setDBValue(db, parts[1], strings.TrimPrefix(message, parts[0]+" "+parts[1]+" "))
			if err == nil {
				addReaction(msg, emojiCommandOK, rtm)
			} else {
				addReaction(msg, emojiCommandError, rtm)
			}
		}
	case "get":
		if len(parts) < 2 {
			addReaction(msg, emojiParametersWrong, rtm)
			return ""
		} else {
			data, err := getDBValue(db, parts[1])
			if err == nil {
				returnMessage = data
			} else {
				addReaction(msg, emojiCommandWarning, rtm)
			}
		}
	case "shark":
		go sharkProc(msg.Channel, rtm, 30, 2)
	case "animate":
		go animateProc(msg.Channel, rtm, 30)
	default:

	}

	if returnMessage != "" {
		returnMessage = prefix + returnMessage
	}

	return returnMessage
}
