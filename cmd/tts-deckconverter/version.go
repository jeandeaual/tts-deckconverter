package main

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var (
	version      string
	buildTimeStr string
	buildTime    time.Time
)

var sha1Regex = regexp.MustCompile("[a-f0-9]{40}")

func init() {
	var err error

	if len(buildTimeStr) > 0 {
		buildTime, err = time.Parse("2006-01-02T15:04:05", buildTimeStr)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			os.Exit(1)
		}
	}
}

func isSHA1(str string) bool {
	return sha1Regex.MatchString(str)
}

func getGoVersion() string {
	return strings.TrimPrefix(runtime.Version(), "go")
}

func displayBuildInformation() {
	if isSHA1(version) {
		fmt.Printf("tts-deckconverter version %s\n", version[:7])
	} else {
		fmt.Printf("tts-deckconverter version %s\n", version)
	}
	if !buildTime.IsZero() {
		fmt.Printf("Built with Go version %s on %s\n", getGoVersion(), buildTime.Format(time.RFC3339))
	} else {
		fmt.Printf("Built with Go version %s\n", getGoVersion())
	}
}
