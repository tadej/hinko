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

// Package ascii contains animation and image2ascii conversion experiments
package ascii

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	asciicanvas "github.com/tompng/go-ascii-canvas"
)

// InitAnimation is called when the ASCII animation's first frame is set up
type InitAnimation func(string) (string, string)

// UpdateAnimation is called for each subsequent frame
type UpdateAnimation func(string, string, string)

// DoSharkAnimation iterates through a one-line shark animation. Updates are handled through initialCall and subsequentCall functions
func DoSharkAnimation(len int, maxTurns int, delay int64, initialCall InitAnimation, subsequentCalls UpdateAnimation) {
	var shark string
	var right bool

	right = false
	shark = getSharkString(0, len, right)

	channel, timestamp := initialCall(shark)

	for turns := 0; turns < maxTurns; turns++ {
		right = !right
		for i := 1; i < len-1; i++ {
			time.Sleep(time.Duration(delay) * time.Millisecond)
			var newMsg string
			if right {
				newMsg = getSharkString(i, len, right)
			} else {
				newMsg = getSharkString(len-i, len, right)
			}
			subsequentCalls(channel, timestamp, newMsg)
		}
	}
	newMsg := getSharkString(-1, len, right)
	subsequentCalls(channel, timestamp, newMsg)
}

// DoFrameAnimation iterates through frames of a string of ASCII pictures. Updates are handled through initialCall and subsequentCall functions
func DoFrameAnimation(len int, delay int64, initialCall InitAnimation, subsequentCalls UpdateAnimation) {
	var anim string
	anim = getAnimationFrame(0)
	channel, timestamp := initialCall(anim)

	for i := 1; i < len; i++ {
		time.Sleep(time.Duration(delay) * time.Millisecond)
		var newMsg string
		newMsg = getAnimationFrame(i)
		subsequentCalls(channel, timestamp, newMsg)
	}
}

func getAnimationFrame(i int) string {
	frames := [4]string{
		"```╔════╤╤╤╤════╗\n" +
			"║    │││ \\   ║\n" +
			"║    │││  O  ║\n" +
			"║    OOO     ║```",
		"```╔════╤╤╤╤════╗\n" +
			"║    ││││    ║\n" +
			"║    ││││    ║\n" +
			"║    OOOO    ║```",
		"```╔════╤╤╤╤════╗\n" +
			"║   / │││    ║\n" +
			"║  O  │││    ║\n" +
			"║     OOO    ║```",
		"```╔════╤╤╤╤════╗\n" +
			"║    ││││    ║\n" +
			"║    ││││    ║\n" +
			"║    OOOO    ║```"}

	return frames[i%len(frames)]
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

	return ":shark: `" + ret + "`"
}

func clampLongerRectangleSize(w float64, h float64, maxLonger float64, maxShorter float64) (float64, float64) {
	ratio := h / w
	var w1 float64
	var h1 float64

	if w > h {
		w1 = maxLonger
		h1 = ratio * w1
		if h1 > maxShorter {
			h1 = maxShorter
			w1 = h1 / ratio
		}
	} else {
		h1 = maxLonger
		w1 = h1 / ratio
		if w1 > maxShorter {
			w1 = maxShorter
			h1 = w1 * ratio
		}
	}

	return w1, h1
}

func getXYCorrelatedWithAB(x int, y int, A float64, B float64) (int, int) {
	if A > B {
		return x, y
	}
	return y, x
}

// ImageToASCII downloads an image from a URL and returns a code-formatted (``) ASCII representation
func ImageToASCII(url string) (string, error) {

	// TG: determined by experimentation that Slack doesn't allow code segments larger than 100x70 or 70x100
	var slackWidthLimit float64 = 100
	var slackHeightLimit float64 = 70

	var err error

	url = strings.Trim(url, "<>")
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	img, err := asciicanvas.NewImageBufferFromBytes(bytes)
	if err != nil {
		return "", err
	}

	width, height := clampLongerRectangleSize(float64(img.Width), float64(img.Height), slackWidthLimit, slackHeightLimit)

	bufferW, bufferH := getXYCorrelatedWithAB(int(slackWidthLimit), int(slackHeightLimit), width, height)

	canvas := asciicanvas.NewImageBuffer(bufferW, bufferH)

	canvas.Draw(img, 0, 0, width, height)

	str := canvas.String()
	str = strings.Replace(str, "`", "'", -1)
	fmt.Printf(str)

	return "```" + str + "```", err
}
