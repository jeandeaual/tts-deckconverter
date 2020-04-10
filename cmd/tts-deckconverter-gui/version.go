package main

import (
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

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
