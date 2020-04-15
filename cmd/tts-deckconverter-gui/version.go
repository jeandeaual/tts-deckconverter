package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
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
	if len(buildTimeStr) > 0 {
		var err error

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

func getModuleVersion(modulePath string) (string, error) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "", errors.New("couldn't retrieve build information")
	}
	for _, module := range bi.Deps {
		if module.Path == modulePath {
			return strings.TrimLeft(module.Version, "v"), nil
		}
	}

	return "", fmt.Errorf("couldn't find %s in the build dependencies", modulePath)
}
