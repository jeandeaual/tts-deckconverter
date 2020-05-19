package custom

import (
	"github.com/jeandeaual/tts-deckconverter/plugins"
)

type customPlugin struct {
	id   string
	name string
}

func (p customPlugin) PluginID() string {
	return p.id
}

func (p customPlugin) PluginName() string {
	return p.name
}

func (p customPlugin) AvailableOptions() plugins.Options {
	return plugins.Options{}
}

func (p customPlugin) URLHandlers() []plugins.URLHandler {
	return []plugins.URLHandler{}
}

func (p customPlugin) FileExtHandlers() map[string]plugins.FileHandler {
	return map[string]plugins.FileHandler{}
}

func (p customPlugin) DeckTypeHandlers() map[string]plugins.FileHandler {
	return map[string]plugins.FileHandler{}
}

func (p customPlugin) GenericFileHandler() plugins.FileHandler {
	return fromList
}

func (p customPlugin) AvailableBacks() map[string]plugins.Back {
	return map[string]plugins.Back{}
}

// CustomPlugin is the exported plugin for this package
var CustomPlugin = customPlugin{
	id:   "custom",
	name: "Custom",
}
