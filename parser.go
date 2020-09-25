package deckconverter

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/jeandeaual/tts-deckconverter/log"
	"github.com/jeandeaual/tts-deckconverter/plugins"
)

func parseFileWithPlugin(target string, plugin plugins.Plugin, options map[string]string) ([]*plugins.Deck, error) {
	log.Infof("Parsing file %s", target)

	var decks []*plugins.Deck

	if _, err := os.Stat(target); os.IsNotExist(err) {
		return nil, err
	}

	file, err := os.Open(target)
	if err != nil {
		return nil, err
	}
	defer func() {
		cerr := file.Close()
		if cerr != nil {
			log.Error(cerr)
		}
	}()

	ext := filepath.Ext(target)
	name := strings.TrimSuffix(filepath.Base(target), ext)

	log.Debugf("Base file name: %s", name)

	if handler, ok := plugin.FileExtHandlers()[ext]; ok {
		decks, err = handler(file, name, options)
		return decks, err
	}

	decks, err = plugin.GenericFileHandler().FileHandler(file, name, options)
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
		cerr := file.Close()
		if cerr != nil {
			log.Error(cerr)
		}
	}()

	// Get the name of the file, without the folder and extension
	name := strings.TrimSuffix(filepath.Base(target), ext)

	log.Debugf("Base file name: %s", name)

	decks, err := fileExtHandler(file, name, options)

	return decks, err
}

// Parse a URL or file and generate a list of decks from it.
func Parse(target, mode string, options map[string]string) ([]*plugins.Deck, error) {
	if u, err := url.Parse(target); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		// Check if the target is a supported URL
		for _, handler := range URLHandlers {
			if handler.Regex.MatchString(target) {
				log.Debugf("Using handler %+v", handler)
				decks, err := handler.Handler(target, options)
				return decks, err
			}
		}

		return nil, fmt.Errorf("unsupported URL: %s", target)
	}

	_, err := os.Stat(target)

	if err != nil {
		return nil, fmt.Errorf("file %s not found: %w", target, err)
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
