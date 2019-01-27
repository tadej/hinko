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
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

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
