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
	"strings"
	"time"

	"github.com/nlopes/slack"
)

func main() {

	token := "xoxb-3449240089-534667652646-i0dKFCTxYiiTA0nFY1jqWhbr" //os.Getenv("SLACK_TOKEN")
	api := slack.New(token)
	rtm := api.NewRTM()
	go rtm.ManageConnection()

Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.ConnectedEvent:
				//fmt.Println("Connection counter:", ev.ConnectionCount)

			case *slack.MessageEvent:
				info := rtm.GetInfo()

				user, err := api.GetUserInfo(ev.User)
				if err != nil {
					//fmt.Printf("%s %s\n", ev.User, err)
					continue
				}

				var directMessage = isIMChannel(api, ev.Channel)

				prefix := fmt.Sprintf("<@%s> ", info.User.ID)

				if ev.User != info.User.ID {
					respond(info.User.ID, rtm, ev, prefix, user.ID, directMessage)
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

func respond(botID string, rtm *slack.RTM, msg *slack.MessageEvent, prefix string, user string, directMessage bool) {
	var response string
	text := msg.Text

	text = strings.TrimPrefix(text, prefix)
	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	var mentionedBot = strings.HasPrefix(msg.Text, "<@"+botID+">")

	if directMessage || mentionedBot {
		response = processMessage(text, user, directMessage, msg, rtm)
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

func getAnimation(i int) string {
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

func processMessage(message string, userID string, directMessage bool, msg *slack.MessageEvent, rtm *slack.RTM) string {
	var returnMessage string
	var prefix string

	// mention, tag the user in the channel
	if !directMessage {
		prefix = "<@" + userID + "> "
	}

	if strings.HasPrefix(message, "teams") {

	} else if strings.HasPrefix(message, "animate") {
		var anim string
		anim = getAnimation(0)
		len := 30
		retChan, retTimeStamp, err := rtm.PostMessage(msg.Channel, slack.MsgOptionText(anim, false), slack.MsgOptionAsUser(true))
		if err != nil {
			fmt.Printf("%s\n", err)
			return ""
		}

		for i := 1; i < len; i++ {
			timer := time.NewTimer(100 * time.Millisecond)
			<-timer.C
			var newMsg string
			newMsg = getAnimation(i)
			_, _, _, err = rtm.UpdateMessage(retChan, retTimeStamp, slack.MsgOptionText(newMsg, false), slack.MsgOptionAsUser(true))
			if err != nil {
				fmt.Printf("%s\n", err)
				return ""
			}
		}
	} else if strings.HasPrefix(message, "shark") {
		var shark string
		var right bool

		len := 30
		maxTurns := 4

		right = false
		shark = getSharkString(0, len, right)

		retChan, retTimeStamp, err := rtm.PostMessage(msg.Channel, slack.MsgOptionText(shark, false), slack.MsgOptionAsUser(true))
		if err != nil {
			fmt.Printf("%s\n", err)
			return ""
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
					return ""
				}
			}
		}
		newMsg := getSharkString(-1, len, right)
		_, _, _, err = rtm.UpdateMessage(retChan, retTimeStamp, slack.MsgOptionText(newMsg, false), slack.MsgOptionAsUser(true))
		if err != nil {
			fmt.Printf("%s\n", err)
			return ""
		}
	} else {
		var itemRef slack.ItemRef
		itemRef.Channel = msg.Channel
		itemRef.Timestamp = msg.Timestamp
		err := rtm.AddReaction("shrug", itemRef)
		if err != nil {
			fmt.Printf("%s\n", err)
			return ""
		}
	}

	if returnMessage != "" {
		returnMessage = prefix + returnMessage
	}

	return returnMessage
}
