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
	"errors"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/nlopes/slack"
	"github.com/tadej/hinko/ascii"
	"github.com/tadej/hinko/model"
)

func main() {
	token := os.Getenv("SLACK_TOKEN")
	api := slack.New(token)
	rtm := api.NewRTM()
	go rtm.ManageConnection()
	model.OpenDatabase("/tmp/model.lvl")

	setupCloseHandler()

	fmt.Println("Starting Slack API loop")
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

	defer model.CloseDatabase()
}

// SetupCloseHandler creates a 'listener' on a new goroutine which will notify the
// program if it receives an interrupt from the OS. We then handle this by calling
// our clean up procedure and exiting the program.
func setupCloseHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		model.CloseDatabase()
		os.Exit(0)
	}()
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

	var mentionedBot = strings.HasPrefix(msg.Text, "<@"+botID+">")

	if directMessage || mentionedBot {
		response = processMessage(text, user, directMessage, msg, rtm)
		if response != "" {
			rtm.SendMessage(rtm.NewOutgoingMessage(response, msg.Channel))
		}
	}
}

func sharkProc(channel string, rtm *slack.RTM, len int, maxTurns int) {
	var shark string
	var right bool

	right = false
	shark = ascii.GetSharkString(0, len, right)

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
				newMsg = ascii.GetSharkString(i, len, right)
			} else {
				newMsg = ascii.GetSharkString(len-i, len, right)
			}
			_, _, _, err = rtm.UpdateMessage(retChan, retTimeStamp, slack.MsgOptionText(newMsg, false), slack.MsgOptionAsUser(true))
			if err != nil {
				fmt.Printf("%s\n", err)
			}
		}
	}
	newMsg := ascii.GetSharkString(-1, len, right)
	_, _, _, err = rtm.UpdateMessage(retChan, retTimeStamp, slack.MsgOptionText(newMsg, false), slack.MsgOptionAsUser(true))
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
}

func animateProc(channel string, rtm *slack.RTM, len int) {
	var anim string
	anim = ascii.GetAnimationFrame(0)
	retChan, retTimeStamp, err := rtm.PostMessage(channel, slack.MsgOptionText(anim, false), slack.MsgOptionAsUser(true))
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	for i := 1; i < len; i++ {
		timer := time.NewTimer(100 * time.Millisecond)
		<-timer.C
		var newMsg string
		newMsg = ascii.GetAnimationFrame(i)
		_, _, _, err = rtm.UpdateMessage(retChan, retTimeStamp, slack.MsgOptionText(newMsg, false), slack.MsgOptionAsUser(true))
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
	}
}

func addReaction(msg *slack.MessageEvent, reaction string, rtm *slack.RTM) {
	if msg.Username == "slackbot" {
		return
	}

	var itemRef slack.ItemRef
	itemRef.Channel = msg.Channel
	itemRef.Timestamp = msg.Timestamp

	err := rtm.AddReaction(reaction, itemRef)
	if err != nil {
		fmt.Printf("Adding reaction, %s\n", err)
		return
	}
}

func getInfoMessage() string {
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

func shuffle(vals []string) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for len(vals) > 0 {
		n := len(vals)
		randIndex := r.Intn(n)
		vals[n-1], vals[randIndex] = vals[randIndex], vals[n-1]
		vals = vals[:n-1]
	}
}

func getRandomTeams(teamSize int, members []string, membersCanRepeat bool, teamNames []string, shuffleTeamNames bool) (string, error) {
	ret := "\n"

	if teamSize > len(members)/2 {
		return "", errors.New("Team size can't be more than half the group size")
	}

	shuffle(members)

	if shuffleTeamNames {
		shuffle(teamNames)
	}

	d := len(members) % teamSize

	if d != 0 && membersCanRepeat {
		add := (teamSize - d)
		newMembers := make([]string, len(members)+add)
		copy(newMembers, members)

		if d != 0 && membersCanRepeat {
			for i := 0; i < add; i++ {
				newMembers[len(members)+i] = members[i]
			}
		}
		members = newMembers
	}

	teamIndex := 0
	currentTeamSize := 0

	for i, member := range members {
		if i%teamSize == 0 {
			if i == len(members)-1 {
				ret += "➕ "
			} else {
				teamIndex++
				currentTeamSize = 0
				if teamIndex-1 < len(teamNames) {
					ret += "\n" + teamNames[teamIndex-1] + ": "
				} else {
					ret += "\nTeam " + strconv.Itoa(teamIndex) + ": "
				}
			}
		}
		ret += member + " "
		currentTeamSize++
	}

	if currentTeamSize < teamSize {
		ret += "➕❓"
	}

	return ret, nil
}

func processMessage(message string, userID string, directMessage bool, msg *slack.MessageEvent, rtm *slack.RTM) string {
	var returnMessage string
	var prefix string

	// mention, tag the user in the channel
	if !directMessage {
		prefix = "<@" + userID + ">\n"
	}

	parts := strings.Split(message, " ")

	acceptedCommands := map[string]bool{
		"put":         true,
		"get":         true,
		"shark":       true,
		"animate":     true,
		"help":        true,
		"randompairs": true,
		"randomteams": true,
		"group":       true,
		"ascii":       true,
	}

	emojiCommandNotFound := "shrug"                  // 🤷‍♀️
	emojiParametersWrong := "heavy_multiplication_x" // ✖️
	emojiCommandError := "bug"                       // 🐛
	emojiCommandOK := "heavy_check_mark"             // ✔️
	emojiCommandWarning := "grey_question"           // ❔

	if len(parts) < 1 || !acceptedCommands[parts[0]] {
		addReaction(msg, emojiCommandNotFound, rtm)
		return ""
	}

	switch cmd := parts[0]; cmd {
	case "help":
		returnMessage = getInfoMessage()

	case "ascii":

		returnMessage = ascii.ImageToASCII(parts[1])

	case "group":
		if len(parts) < 3 {
			addReaction(msg, emojiParametersWrong, rtm)
			return ""
		}

		groupName := parts[1]

		switch parts[2] {
		case "list":

		case "create":
			err := model.SetGroup(groupName, parts[3:])
			if err != nil {
				addReaction(msg, emojiParametersWrong, rtm)
				return ""
			}
		case "add":
			err := model.AddToGroup(groupName, parts[3:])
			if err != nil {
				addReaction(msg, emojiParametersWrong, rtm)
				return ""
			}

		case "remove":
			err := model.RemoveFromGroup(groupName, parts[3:])
			if err != nil {
				addReaction(msg, emojiParametersWrong, rtm)
				return ""
			}
		default:
			addReaction(msg, emojiParametersWrong, rtm)
			return ""

		}

		addReaction(msg, emojiCommandOK, rtm)

		groupTest, err := model.GetGroup(parts[1])
		if err != nil {
			addReaction(msg, emojiParametersWrong, rtm)
			return ""
		}
		returnMessage = "`" + parts[1] + "` members: " + strings.Join(groupTest, " ")
	case "randompairs":
		if len(parts) < 2 {
			addReaction(msg, emojiParametersWrong, rtm)
			return ""
		}

		var members []string
		var err error
		// only one parameter means group name
		if len(parts) == 2 {
			members, err = model.GetGroup(parts[1])
			if err != nil {
				addReaction(msg, emojiParametersWrong, rtm)
				return ""
			}
		} else {
			members = parts[1:]
		}

		teamNames, _ := model.GetGroup("pairnames")

		returnMessage, err = getRandomTeams(2, members, true, teamNames, false)
		if err != nil {
			addReaction(msg, emojiParametersWrong, rtm)
			return ""
		}

	case "randomteams":
		if len(parts) < 3 {
			addReaction(msg, emojiParametersWrong, rtm)
			return ""
		}

		teamSize, err := strconv.Atoi(parts[1])
		if err != nil {
			addReaction(msg, emojiParametersWrong, rtm)
			return ""
		}

		var members []string

		// only one parameter means group name
		if len(parts) == 3 {
			members, err = model.GetGroup(parts[2])
			if err != nil {
				addReaction(msg, emojiParametersWrong, rtm)
				return ""
			}
		} else {
			members = parts[2:]
		}

		teamNames, _ := model.GetGroup("teamnames")

		returnMessage, err = getRandomTeams(teamSize, members, false, teamNames, true)
		if err != nil {
			addReaction(msg, emojiParametersWrong, rtm)
			return ""
		}

	case "put":
		if len(parts) < 3 {
			addReaction(msg, emojiParametersWrong, rtm)
			return ""
		}

		err := model.SetDBValue(parts[1], strings.TrimPrefix(message, parts[0]+" "+parts[1]+" "))
		if err == nil {
			addReaction(msg, emojiCommandOK, rtm)
		} else {
			addReaction(msg, emojiCommandError, rtm)
		}

	case "get":
		if len(parts) < 2 {
			addReaction(msg, emojiParametersWrong, rtm)
			return ""
		}

		data, err := model.GetDBValue(parts[1])
		if err == nil {
			returnMessage = data
		} else {
			addReaction(msg, emojiCommandWarning, rtm)
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
