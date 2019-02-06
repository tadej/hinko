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

// Package commands includes all the command and subcommand processing
package commands

import (
	"errors"
	"strconv"
	"strings"

	"github.com/tadej/hinko/ascii"
	"github.com/tadej/hinko/model"
	"github.com/tadej/hinko/slack"
)

// ProcessCommand - function signature — all command processing functions adhere to this format
type ProcessCommand func([]string, slack.MessageInfo) string

// AcceptedCommands - accepted text commands with their corresponding processor functions
var AcceptedCommands = map[string]ProcessCommand{
	"put":         ProcessCommandPut,
	"get":         ProcessCommandGet,
	"shark":       ProcessCommandShark,
	"animate":     ProcessCommandAnimate,
	"help":        ProcessCommandHelp,
	"randompairs": ProcessCommandRandomPairs,
	"randomteams": ProcessCommandRandomTeams,
	"group":       ProcessCommandGroup,
	"ascii":       ProcessCommandASCII,
	"score":       ProcessCommandScore,
}

// AcceptedGroupSubCommands - accepted text group subcommands with their corresponding processor functions
var AcceptedGroupSubCommands = map[string]ProcessCommand{
	"set":    ProcessCommandGroupSet,
	"create": ProcessCommandGroupSet,
	"add":    ProcessCommandGroupAdd,
	"remove": ProcessCommandGroupRemove,
	"list":   ProcessCommandGroupList,
}

// AcceptedScoreSubCommands - accepted score subcommands with their corresponding processor functions
var AcceptedScoreSubCommands = map[string]ProcessCommand{
	"set":   ProcessCommandScoreSet,
	"add":   ProcessCommandScoreSet,
	"get":   ProcessCommandScoreGet,
	"reset": ProcessCommandScoreReset,
}

// EmojiCommandNotFound ‍‍🤷‍♀️
var EmojiCommandNotFound = "shrug"

// EmojiParametersWrong ✖️
var EmojiParametersWrong = "heavy_multiplication_x"

// EmojiCommandError 🐛
var EmojiCommandError = "bug"

// EmojiCommandOK ✔️
var EmojiCommandOK = "heavy_check_mark"

// EmojiCommandWarning ❔
var EmojiCommandWarning = "grey_question"

// TeamNamesGroup DB key contains a list of space-delimited team names
var TeamNamesGroup = "teamnames"

// PairNamesGroup DB key contains a list of space-delimited pair names
var PairNamesGroup = "pairnames"

// ProcessCommandHelp returns a help message
func ProcessCommandHelp(parts []string, msg slack.MessageInfo) string {
	infoMessage :=
		`Try the following commands:` + "\n" +
			"`help`\n" +
			"`put key value`\n" +
			"`get key`\n" +
			"`group groupname list`\n" +
			"`group groupname create @user1 @user2 @user3 ...`\n" +
			"`group groupname add @user1 @user2 @user3 ...`\n" +
			"`group groupname remove @user1 @user2 @user3 ...`\n" +
			"`score add team1:team2 score1:score2`\n" +
			"`score get team2:team1`\n" +
			"`score reset team2:team1`\n" +
			"`score reset team2:team1 score2:score1`\n" +
			"`randompairs @user1 @user2 @user3 ...`\n" +
			"`randompairs group`\n" +
			"`randomteams teamsize @user1 @user2 @user3 ...`\n" +
			"`randomteams teamsize group`\n" +
			"`ascii https://imageurl`\n" +
			"`shark`\n" +
			"`animate`\n\n" +
			" reserved groups: _pairnames_, _teamnames_\n\n" +
			"More info:\nhttps://github.com/tadej/hinko"

	return infoMessage
}

// ProcessCommandASCII converts an image to ASCII
func ProcessCommandASCII(parts []string, msg slack.MessageInfo) string {
	ret, err := ascii.ImageToASCII(parts[1])
	if err != nil {
		React(msg, EmojiCommandError)
		return ""
	}
	return ret
}

// ProcessCommandShark animates an ASCII shark
func ProcessCommandShark(parts []string, msg slack.MessageInfo) string {
	// sending functon bodies as parameter so ascii doesn't have to know "slack"
	go ascii.DoSharkAnimation(30, 2, 300,
		func(txt string) (string, string) {
			channel, timestamp, _ := slack.PostMessage(msg.Channel, txt)
			return channel, timestamp
		}, func(channel string, timestamp string, newTxt string) {
			_ = slack.UpdateMessage(channel, timestamp, newTxt)
		})

	return ""
}

// ProcessCommandAnimate animates a pendulum in ASCII
func ProcessCommandAnimate(parts []string, msg slack.MessageInfo) string {
	// sending functon bodies as parameter so ascii doesn't have to know "slack"
	go ascii.DoFrameAnimation(30, 300,
		func(txt string) (string, string) {
			channel, timestamp, _ := slack.PostMessage(msg.Channel, txt)
			return channel, timestamp
		}, func(channel string, timestamp string, newTxt string) {
			_ = slack.UpdateMessage(channel, timestamp, newTxt)
		})

	return ""
}

// ProcessCommandGroupList lists members of a group
func ProcessCommandGroupList(parts []string, msg slack.MessageInfo) string {
	group, err := model.GetGroup(parts[1])
	if err != nil {
		React(msg, EmojiCommandWarning)
		return ""
	}
	return "`" + parts[1] + "` members: " + strings.Join(group, " ")
}

// ProcessCommandGroupSet creates a new group
func ProcessCommandGroupSet(parts []string, msg slack.MessageInfo) string {
	ProcessGroupCommandError(model.SetGroup(parts[1], parts[3:]), msg, true)
	return ""
}

// ProcessCommandGroupAdd adds members to a group
func ProcessCommandGroupAdd(parts []string, msg slack.MessageInfo) string {
	ProcessGroupCommandError(model.AddToGroup(parts[1], parts[3:]), msg, true)
	return ""
}

// ProcessCommandGroupRemove removes members from a group
func ProcessCommandGroupRemove(parts []string, msg slack.MessageInfo) string {
	ProcessGroupCommandError(model.RemoveFromGroup(parts[1], parts[3:]), msg, true)
	return ""
}

// ProcessGroupCommandError processes errors Reactions for group commands
func ProcessGroupCommandError(err error, msg slack.MessageInfo, confirmSuccess bool) {
	if err != nil {
		React(msg, EmojiCommandWarning)
	} else if confirmSuccess {
		React(msg, EmojiCommandOK)
	}
}

// ProcessCommandGroup decides which group processing function to call
func ProcessCommandGroup(parts []string, msg slack.MessageInfo) string {
	if len(parts) < 3 {
		React(msg, EmojiParametersWrong)
		return ""
	}

	fn := AcceptedGroupSubCommands[strings.ToLower(parts[2])]

	if fn != nil {
		return fn(parts, msg)
	}

	React(msg, EmojiCommandError)
	return ""
}

// ProcessCommandScore decides which score processing function to call
func ProcessCommandScore(parts []string, msg slack.MessageInfo) string {
	if len(parts) < 3 {
		React(msg, EmojiParametersWrong)
		return ""
	}

	fn := AcceptedScoreSubCommands[strings.ToLower(parts[1])]

	if fn != nil {
		return fn(parts, msg)
	}

	React(msg, EmojiCommandError)
	return ""
}

func teamsNamesFromString(input string) (string, string, error) {
	input = strings.ToUpper(input)
	str := strings.Split(input, ":")
	if len(str) != 2 {
		return "", "", errors.New("can't get team names from string")
	}

	return str[0], str[1], nil
}

func scoresFromString(input string) (int, int, error) {
	str := strings.Split(input, ":")
	if len(str) != 2 {
		return 0, 0, errors.New("can't get team names from string")
	}

	sc1, err := strconv.Atoi(str[0])

	if err != nil {
		return 0, 0, err
	}

	sc2, err := strconv.Atoi(str[1])

	if err != nil {
		return 0, 0, err
	}

	return sc1, sc2, nil
}

// ProcessCommandScoreSet sets a new score with score set team1:team2 score1:score2
func ProcessCommandScoreSet(parts []string, msg slack.MessageInfo) string {
	if len(parts) < 4 {
		React(msg, EmojiParametersWrong)
		return ""
	}

	team1, team2, err1 := teamsNamesFromString(parts[2])
	score1, score2, err2 := scoresFromString(parts[3])
	var err error
	if err1 == nil && err2 == nil {
		err = model.AddScore(team1, team2, score1, score2)
		if err == nil {
			React(msg, EmojiCommandOK)
			return ""
		}
	}

	React(msg, EmojiParametersWrong)

	return ""
}

// ProcessCommandScoreReset resets the score for team1:team2
func ProcessCommandScoreReset(parts []string, msg slack.MessageInfo) string {
	team1, team2, err := teamsNamesFromString(parts[2])

	score1 := 0
	score2 := 0

	if err == nil {
		if len(parts) >= 4 {
			score1, score2, err = scoresFromString(parts[3])
		}

		err = model.ResetScore(team1, team2, score1, score2)
		if err == nil {
			React(msg, EmojiCommandOK)
			return ""
		}
	}

	React(msg, EmojiParametersWrong)

	return ""
}

// ProcessCommandScoreGet gets the scores for team1:team2
func ProcessCommandScoreGet(parts []string, msg slack.MessageInfo) string {
	team1, team2, err1 := teamsNamesFromString(parts[2])

	if err1 == nil {
		scoreInfo, err := model.GetScores(team1, team2)

		if err != nil {
			React(msg, EmojiParametersWrong)
		} else {
			return model.GetScoreMessage(scoreInfo)
		}
	} else {
		React(msg, EmojiParametersWrong)
	}

	return ""
}

// ProcessCommandRandomPairs assembles random pairs
func ProcessCommandRandomPairs(parts []string, msg slack.MessageInfo) string {
	var returnMessage string

	if len(parts) < 2 {
		React(msg, EmojiParametersWrong)
		return ""
	}

	members, err := getReferencedMembers(parts, 1)
	if err != nil {
		React(msg, EmojiParametersWrong)
		return ""
	}

	teamNames, _ := model.GetGroup(PairNamesGroup)

	returnMessage, err = model.GetRandomTeams(2, members, true, teamNames, false)
	if err != nil {
		React(msg, EmojiParametersWrong)
		return ""
	}

	return returnMessage
}

// ProcessCommandRandomTeams assembles random teams
func ProcessCommandRandomTeams(parts []string, msg slack.MessageInfo) string {
	var returnMessage string

	if len(parts) < 3 {
		React(msg, EmojiParametersWrong)
		return ""
	}

	teamSize, err := strconv.Atoi(parts[1])
	if err != nil {
		React(msg, EmojiParametersWrong)
		return ""
	}

	members, err := getReferencedMembers(parts, 2)
	if err != nil {
		React(msg, EmojiParametersWrong)
		return ""
	}

	teamNames, _ := model.GetGroup(TeamNamesGroup)

	returnMessage, err = model.GetRandomTeams(teamSize, members, false, teamNames, true)
	if err != nil {
		React(msg, EmojiParametersWrong)
		return ""
	}

	return returnMessage
}

// ProcessCommandPut puts value at key
func ProcessCommandPut(parts []string, msg slack.MessageInfo) string {
	if len(parts) < 3 {
		React(msg, EmojiParametersWrong)
		return ""
	}

	err := model.SetDBValue(parts[1], strings.Join(parts[2:], " "))
	if err == nil {
		React(msg, EmojiCommandOK)
	} else {
		React(msg, EmojiCommandError)
	}

	return ""
}

// ProcessCommandGet gets value at key
func ProcessCommandGet(parts []string, msg slack.MessageInfo) string {
	var returnMessage string

	if len(parts) < 2 {
		React(msg, EmojiParametersWrong)
		return ""
	}

	data, err := model.GetDBValue(parts[1])
	if err == nil {
		returnMessage = data
	} else {
		React(msg, EmojiCommandWarning)
	}
	return returnMessage
}

func getReferencedMembers(parts []string, offset int) ([]string, error) {
	var members []string
	var err error

	// only one parameter means we treat it as a group name
	if len(parts) == offset+1 {
		members, err = model.GetGroup(parts[offset])
	} else {
		members = parts[offset:]
	}

	return members, err
}

// React adds Slack Reaction (Emoji)
func React(msg slack.MessageInfo, Reaction string) {
	slack.AddReaction(msg.Username, msg.Channel, msg.Timestamp, Reaction)
}
