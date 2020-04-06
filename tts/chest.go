package tts

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"deckconverter/log"
)

// FindChestPath tries to the find TableTop Simulator check folder
// (where the saved objects are located).
func FindChestPath() (string, error) {
	var chestPath string

	switch runtime.GOOS {
	case "windows":
		home := os.Getenv("USERPROFILE")
		if len(home) == 0 {
			return chestPath, errors.New("%USERPROFILE% is not set")
		}
		chestPath = filepath.Join(home, "/Documents/My Games/Tabletop Simulator/Saves/Saved Objects")
	default:
		home := os.Getenv("HOME")
		if len(home) == 0 {
			return chestPath, errors.New("$HOME is not set")
		}
		chestPath = filepath.Join(home, "/.local/share/Tabletop Simulator/Saves/Saved Objects")
	}

	log.Debugf("Chest path: \"%s\"", chestPath)

	if stat, err := os.Stat(chestPath); os.IsNotExist(err) {
		return chestPath, fmt.Errorf("chest path \"%s\" doesn't exist", chestPath)
	} else if err != nil {
		return chestPath, err
	} else if !stat.IsDir() {
		return chestPath, fmt.Errorf("chest path \"%s\" is not a directory", chestPath)
	}

	return chestPath, nil
}
