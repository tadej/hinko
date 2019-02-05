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

// Package model contains db access and data manipulation functions
package model

import (
	"encoding/json"
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// ScoreInfo struct keeps the scores for team1 and team2
type ScoreInfo struct {
	Team1  string
	Team2  string
	Points Score
	Scores []Score
}

// Score struct keeps the score of a current match between team1 and team2
type Score struct {
	Team1     int
	Team2     int
	Timestamp string
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// GetGroup returns a list of members in group name
func GetGroup(name string) ([]string, error) {
	group, err := GetDBValue("[group::" + name + "]")

	if err != nil {
		return nil, err
	}

	members := strings.Split(group, " ")

	return members, nil
}

// SetGroup creates a group with members[]
func SetGroup(name string, members []string) error {
	err := SetDBValue("[group::"+name+"]", strings.Join(members, " "))
	if err != nil {
		return err
	}
	return nil
}

// AddToGroup adds members[] to a group (no duplicates are created)
func AddToGroup(name string, members []string) error {
	var str string

	existingGroup, err := GetGroup(name)
	if err != nil {
		str = ""
	} else {
		str = strings.Join(existingGroup, " ")
		str = strings.Trim(str, " ")
	}

	for _, m := range members {
		if !contains(existingGroup, m) {
			str += " " + m
		}
	}

	str = strings.Trim(str, " ")

	err = SetDBValue("[group::"+name+"]", str)
	if err != nil {
		return err
	}
	return nil
}

// RemoveFromGroup removes members[] if they exist
func RemoveFromGroup(name string, members []string) error {
	existingGroup, err := GetGroup(name)
	if err != nil {
		return err
	}

	str := ""

	for _, m := range existingGroup {
		if !contains(members, m) {
			str += " " + m
		}
	}

	str = strings.Trim(str, " ")

	err = SetDBValue("[group::"+name+"]", str)
	if err != nil {
		return err
	}
	return nil
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

// GetRandomTeams takes a list of strings and puts them into teams of teamSize
func GetRandomTeams(teamSize int, members []string, membersCanRepeat bool, teamNames []string, shuffleTeamNames bool) (string, error) {
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

	// TODO: refactor this a bit. Works but looks messy

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

func reverseScores(scores ScoreInfo) ScoreInfo {
	var retScore ScoreInfo

	retScore.Points.Team1 = scores.Points.Team2
	retScore.Points.Team2 = scores.Points.Team1

	if len(scores.Scores) > 0 {
		retScore.Scores = make([]Score, len(scores.Scores), len(scores.Scores))

		for i := 0; i < len(scores.Scores); i++ {
			retScore.Scores[i].Team1 = scores.Scores[i].Team2
			retScore.Scores[i].Team2 = scores.Scores[i].Team1
		}
	}

	return retScore
}

func scoreInfoToJSON(scoreInfo ScoreInfo) (string, error) {
	js, err := json.Marshal(scoreInfo)
	if err != nil {
		return "", err
	}
	return string(js), nil
}

func jsonToScoreInfo(js string) (ScoreInfo, error) {
	var scoreInfo ScoreInfo
	err := json.Unmarshal([]byte(js), &scoreInfo)
	return scoreInfo, err
}

func getScoreTag(team1 string, team2 string) string {
	return "[SCORE]" + team1 + ":" + team2
}

func orderTeamNames(team1 string, team2 string) (string, string, bool) {
	var t1 string
	var t2 string
	var reverse bool

	t1 = team1
	t2 = team2
	reverse = false

	if strings.Compare(team1, team2) > 0 {
		t2 = team1
		t1 = team2
		reverse = true
	}

	return t1, t2, reverse
}

func addScore(currentScores ScoreInfo, score1 int, score2 int) ScoreInfo {
	var newScores ScoreInfo
	newScores.Team1 = currentScores.Team1
	newScores.Team2 = currentScores.Team2
	newScores.Points = currentScores.Points

	newScores.Scores = make([]Score, len(currentScores.Scores)+1, len(currentScores.Scores)+1)
	copy(newScores.Scores, currentScores.Scores)
	newScores.Scores[len(currentScores.Scores)] = Score{Team1: score1, Team2: score2, Timestamp: time.Now().Format(time.RFC1123)}

	if score1 > score2 {
		newScores.Points.Team1 = currentScores.Points.Team1 + 1
	} else if score2 > score1 {
		newScores.Points.Team2 = currentScores.Points.Team2 + 1
	} else {
		newScores.Points.Team1 = currentScores.Points.Team1 + 1
		newScores.Points.Team2 = currentScores.Points.Team2 + 1
	}

	return newScores
}

// AddScore adds a score for team1 vs team2
func AddScore(team1 string, team2 string, score1 int, score2 int) error {
	var reverse bool
	team1, team2, reverse = orderTeamNames(team1, team2)

	if reverse {
		tmp := score1
		score1 = score2
		score2 = tmp
	}

	scoreInfo, err := GetScores(team1, team2)

	if err != nil {
		scoreInfo.Team1 = team1
		scoreInfo.Team2 = team2
		scoreInfo.Points.Team1 = 0
		scoreInfo.Points.Team2 = 0
	}

	newScores := addScore(scoreInfo, score1, score2)

	js, err := scoreInfoToJSON(newScores)

	tg := getScoreTag(team1, team2)

	if err == nil {
		err = SetDBValue(tg, js)
	}

	return err
}

func scoreToString(score Score) string {
	return strconv.Itoa(score.Team1) + ":" + strconv.Itoa(score.Team2)
}

// GetScoreMessage returns a formatted string describing the score standing
func GetScoreMessage(score ScoreInfo) string {
	var ret string
	ret += "*" + score.Team1 + "*"
	if score.Points.Team1 > score.Points.Team2 {
		ret += " is currently ahead of "
	} else if score.Points.Team1 < score.Points.Team2 {
		ret += " is currently trailing behind "
	} else {
		ret += " is currently tied with "
	}
	ret += "*" + score.Team2 + "*"
	ret += " with " + scoreToString(score.Points) + "."
	ret += "\nHere are the latest standings:"

	for i := 0; i < len(score.Scores); i++ {
		timestr := ""
		t, err := time.Parse(time.RFC1123, score.Scores[i].Timestamp)
		if err == nil {
			timestr = " on " + t.Format(time.RFC1123)
		}
		ret += "\n`" + scoreToString(score.Scores[i]) + timestr + "`"
	}

	return ret
}

// ResetScore resets the score for TEAM1:TEAM2 or TEAM1:TEAM2
func ResetScore(team1 string, team2 string) error {
	team1, team2, _ = orderTeamNames(team1, team2)
	err := SetDBValue(getScoreTag(team1, team2), "")
	return err
}

// GetScores returns an object with the current scores for TEAM1:TEAM2 or TEAM1:TEAM2
func GetScores(team1 string, team2 string) (ScoreInfo, error) {
	var scoreInfo ScoreInfo
	team1, team2, _ = orderTeamNames(team1, team2)

	tag := getScoreTag(team1, team2)
	js, err := GetDBValue(tag)

	if err == nil {
		scoreInfo, err = jsonToScoreInfo(js)
	}

	return scoreInfo, err
}
