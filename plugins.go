package deckconverter

import (
	"log"

	"github.com/jeandeaual/tts-deckconverter/plugins"
	"github.com/jeandeaual/tts-deckconverter/plugins/mtg"
	"github.com/jeandeaual/tts-deckconverter/plugins/pkm"
	"github.com/jeandeaual/tts-deckconverter/plugins/vanguard"
	"github.com/jeandeaual/tts-deckconverter/plugins/ygo"
)

func init() {
	Plugins = make(map[string]plugins.Plugin)
	FileExtHandlers = make(map[string]plugins.FileHandler)

	registerPlugins(
		mtg.MagicPlugin,
		pkm.PokemonPlugin,
		ygo.YGOPlugin,
		vanguard.VanguardPlugin,
	)

	registerURLHandlers()
	registerFileExtHandlers()
}

// Plugins is the list of registered plugins.
var Plugins map[string]plugins.Plugin

// pluginIDs is the ordered list of registered plugins.
var pluginIDs []string

// URLHandlers are all the registered URL handlers.
var URLHandlers []plugins.URLHandler

// FileExtHandlers are all the registered file extension handlers.
var FileExtHandlers map[string]plugins.FileHandler

func registerPlugins(plugins ...plugins.Plugin) {
	for _, plugin := range plugins {
		Plugins[plugin.PluginID()] = plugin
		pluginIDs = append(pluginIDs, plugin.PluginID())
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
	return pluginIDs
}
