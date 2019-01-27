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

import "strings"

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
