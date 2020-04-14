package deckconverter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jeandeaual/tts-deckconverter/log"
	"github.com/jeandeaual/tts-deckconverter/plugins"
)

func parseFileWithPlugin(target string, plugin plugins.Plugin, options map[string]string) ([]*plugins.Deck, error) {
	log.Infof("Parsing file %s", target)
	decks, err := plugin.GenericFileHandler()(target, options)
	return decks, err
}

func parseFile(target string, options map[string]string) ([]*plugins.Deck, error) {
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return nil, err
	}

	// No mode selected, check the file extension handlers
	ext := filepath.Ext(target)

	fileExtHandler, found := FileExtHandlers[ext]
	if !found {
		return nil, fmt.Errorf("no handler found for %s files", ext)
	}

	file, err := os.Open(target)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	// Get the name of the file, without the folder and extension
	name := strings.TrimSuffix(filepath.Base(target), ext)

	log.Debugf("Base file name: %s", name)

	decks, err := fileExtHandler(file, name, options)

	return decks, err
}

// Parse a file or URL and generate a list of decks from it.
func Parse(target, mode string, options map[string]string) ([]*plugins.Deck, error) {
	// Check if the target is a supported URL
	for _, handler := range URLHandlers {
		if handler.Regex.MatchString(target) {
			log.Debugf("Using handler %+v", handler)
			decks, err := handler.Handler(target, options)
			return decks, err
		}
	}

	_, err := os.Stat(target)

	if err != nil {
		return nil, err
	}

	var selectedPlugin *plugins.Plugin

	if len(mode) > 0 {
		plugin, found := Plugins[mode]
		if !found {
			return nil, fmt.Errorf("plugin %s not found", mode)
		}

		log.Infof("Using mode %s", mode)

		selectedPlugin = &plugin
	}

	if selectedPlugin != nil {
		return parseFileWithPlugin(target, *selectedPlugin, options)
	}

	return parseFile(target, options)
}
