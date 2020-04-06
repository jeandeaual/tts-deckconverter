package deckconverter

import (
	"log"
	"sort"

	"deckconverter/plugins"
	"deckconverter/plugins/mtg"
	"deckconverter/plugins/pkm"
	"deckconverter/plugins/ygo"
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

var Plugins map[string]plugins.Plugin
var URLHandlers []plugins.URLHandler
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

func AvailablePlugins() []string {
	keys := make([]string, 0, len(Plugins))
	for key := range Plugins {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
