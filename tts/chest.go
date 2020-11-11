package tts

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jeandeaual/tts-deckconverter/log"
)

// FindChestPath tries to the find TableTop Simulator check folder
// (where the saved objects are located).
func FindChestPath() (string, error) {
	var chestPath string

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		// On some Windows machines `os.UserHomeDir()` seems to return the OneDrive folder
		home = strings.TrimSuffix(home, `\OneDrive`)
		chestPath = filepath.Join(home, `\Documents\My Games\Tabletop Simulator\Saves\Saved Objects`)
	case "darwin":
		chestPath = filepath.Join(home, "/Library/Tabletop Simulator/Saves/Saved Objects")
	default:
		chestPath = filepath.Join(home, "/.local/share/Tabletop Simulator/Saves/Saved Objects")
	}

	log.Debugf("Chest path: \"%s\"", chestPath)

	if stat, err := os.Stat(chestPath); os.IsNotExist(err) {
		return "", fmt.Errorf("chest path \"%s\" doesn't exist", chestPath)
	} else if err != nil {
		return "", err
	} else if !stat.IsDir() {
		return "", fmt.Errorf("chest path \"%s\" is not a directory", chestPath)
	}

	return chestPath, nil
}
