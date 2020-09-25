package pkm

import (
	"github.com/jeandeaual/tts-deckconverter/plugins"
)

type imageQuality string

const (
	normal imageQuality = "normal"
	hires  imageQuality = "hires"
)

type pokemonPlugin struct {
	id   string
	name string
}

func (p pokemonPlugin) PluginID() string {
	return p.id
}

func (p pokemonPlugin) PluginName() string {
	return p.name
}

func (p pokemonPlugin) AvailableOptions() plugins.Options {
	return plugins.Options{
		"quality": plugins.Option{
			Type:        plugins.OptionTypeEnum,
			Description: "image quality",
			AllowedValues: []string{
				string(normal),
				string(hires),
			},
			DefaultValue: string(hires),
		},
	}
}

func (p pokemonPlugin) URLHandlers() []plugins.URLHandler {
	return []plugins.URLHandler{}
}

func (p pokemonPlugin) FileExtHandlers() map[string]plugins.FileHandler {
	return map[string]plugins.FileHandler{
		".ptcgo": fromDeckFile,
	}
}

func (p pokemonPlugin) DeckTypeHandlers() map[string]plugins.DeckType {
	return map[string]plugins.DeckType{}
}

func (p pokemonPlugin) GenericFileHandler() plugins.DeckType {
	return plugins.DeckType{
		FileHandler: fromDeckFile,
		Example: `Pokemon - 11
2 Lucario & Melmetal-GX SM9b 29
2 Hoopa SLG 55
2 Genesect SM9b 36
1 Unown LOT 91
1 Girafarig LOT 94
1 Xurkitree-GX PR-SM SM68
1 Solgaleo Prism Star UPR 89

Trainer - 46
4 Steven's Resolve CES 145
4 Erika's Hospitality TEU 140
3 Lusamine CIN 96
3 Cynthia UPR 119
3 Acerola BUS 112
3 Plumeria BUS 120

Energy - 3
2 Double Colorless Energy SUM 136
1 Rainbow Energy SUM 137
`,
	}
}

func (p pokemonPlugin) AvailableBacks() map[string]plugins.Back {
	return map[string]plugins.Back{
		plugins.DefaultBackKey: {
			URL:         defaultBackURL,
			Description: "standard paper card back",
		},
		"japanese": {
			URL:         japaneseBackURL,
			Description: "Japanese card back",
		},
		"japanese_old": {
			URL:         japaneseOldBackURL,
			Description: "Old Japanese card back",
		},
	}
}

// PokemonPlugin is the exported plugin for this package
var PokemonPlugin = pokemonPlugin{
	id:   "pkm",
	name: "Pokémon",
}
