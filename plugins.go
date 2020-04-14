package deckconverter

import (
	"log"
	"sort"

	"github.com/jeandeaual/tts-deckconverter/plugins"
	"github.com/jeandeaual/tts-deckconverter/plugins/mtg"
	"github.com/jeandeaual/tts-deckconverter/plugins/pkm"
	"github.com/jeandeaual/tts-deckconverter/plugins/ygo"
)

func init() {
	Plugins = make(map[string]plugins.Plugin)
	FileExtHandlers = make(map[string]plugins.FileHandler)

	registerPlugins(
		mtg.MagicPlugin,
		pkm.PokemonPlugin,
		ygo.YGOPlugin,
	)

	registerURLHandlers()
	registerFileExtHandlers()
}

// Plugins is the list of registered deckconverter plugins.
var Plugins map[string]plugins.Plugin

// URLHandlers are all the registered URL handlers.
var URLHandlers []plugins.URLHandler

// FileExtHandlers are all the registered file extension handlers.
var FileExtHandlers map[string]plugins.FileHandler

func registerPlugins(plugins ...plugins.Plugin) {
	for _, plugin := range plugins {
		Plugins[plugin.PluginID()] = plugin
	}
}

func registerURLHandlers() {
	for _, plugin := range Plugins {
		URLHandlers = append(URLHandlers, plugin.URLHandlers()...)
	}
}

func registerFileExtHandlers() {
	for _, plugin := range Plugins {
		for ext, fileExtHandler := range plugin.FileExtHandlers() {
			_, found := FileExtHandlers[ext]
			if found {
				log.Fatalf(
					"Handler for file extension %s already exists, cannot "+
						"register for %s",
					ext,
					plugin.PluginID(),
				)
			}

			FileExtHandlers[ext] = fileExtHandler
		}
	}
}

// AvailablePlugins lists the registered plugins, sorted.
func AvailablePlugins() []string {
	keys := make([]string, 0, len(Plugins))
	for key := range Plugins {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
