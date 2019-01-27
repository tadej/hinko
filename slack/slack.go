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

// Thanks to Spencer Smith for the Slack/GO tutorial: https://rsmitty.github.io/Slack-Bot/
// Thanks https://github.com/nlopes/slack

// Package slack contains everything needed to use the slack API
package slack

import (
	"fmt"

	"github.com/nlopes/slack"
)

var api *slack.Client
var rtm *slack.RTM

// MessageInfo struct that is sent through the message loop channel
type MessageInfo struct {
	OK        bool
	UserID    string
	Username  string
	Channel   string
	IM        bool
	Message   string
	Prefix    string
	MyID      string
	Timestamp string
}

// Init initializes the Slack connection
func Init(token string) {
	api = slack.New(token)
	rtm = api.NewRTM()
	go rtm.ManageConnection()
}

// MessageLoop Loops through incoming Slack messages
func MessageLoop(c chan MessageInfo) {
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

				ret := MessageInfo{OK: true, UserID: user.ID, MyID: info.User.ID,
					Channel: ev.Channel, Prefix: fmt.Sprintf("<@%s> ", info.User.ID),
					IM: isIMChannel(ev.Channel), Message: ev.Text, Username: ev.Username,
					Timestamp: ev.Timestamp}

				if ev.User != info.User.ID {
					c <- ret
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
	c <- MessageInfo{OK: false}
}

// SendMessage sends a message in the selected Slack channel
func SendMessage(channel string, text string) {
	rtm.SendMessage(rtm.NewOutgoingMessage(text, channel))
}

// PostMessage sends a message in the selected Slack channel
func PostMessage(channel string, text string) (string, string, error) {
	retChan, retTimeStamp, err := rtm.PostMessage(channel, slack.MsgOptionText(text, false), slack.MsgOptionAsUser(true))
	return retChan, retTimeStamp, err
}

// UpdateMessage changes the text of an existing message, finding it by channel and timestamp
func UpdateMessage(channel string, timestamp string, text string) error {
	_, _, _, err := rtm.UpdateMessage(channel, timestamp, slack.MsgOptionText(text, false), slack.MsgOptionAsUser(true))
	return err
}

// AddReaction adds the specified reaction to a message defined by channel and timestamp
func AddReaction(author string, channel string, timestamp string, reaction string) {
	if author == "slackbot" {
		return
	}

	var itemRef slack.ItemRef
	itemRef.Channel = channel
	itemRef.Timestamp = timestamp

	err := rtm.AddReaction(reaction, itemRef)
	if err != nil {
		fmt.Printf("Adding reaction, %s\n", err)
		return
	}
}

func isIMChannel(channel string) bool {
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
