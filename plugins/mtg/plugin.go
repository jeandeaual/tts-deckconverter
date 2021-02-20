package mtg

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"

	"github.com/jeandeaual/tts-deckconverter/log"
	"github.com/jeandeaual/tts-deckconverter/plugins"
)

type imageQuality string

const (
	small  imageQuality = "small"
	normal imageQuality = "normal"
	large  imageQuality = "large"
	png    imageQuality = "png"
)

type magicPlugin struct {
	id   string
	name string
}

func (p magicPlugin) PluginID() string {
	return p.id
}

func (p magicPlugin) PluginName() string {
	return p.name
}

func (p magicPlugin) AvailableOptions() plugins.Options {
	return plugins.Options{
		"quality": plugins.Option{
			Type:        plugins.OptionTypeEnum,
			Description: "image quality",
			AllowedValues: []string{
				string(small),
				string(normal),
				string(large),
				string(png),
			},
			DefaultValue: string(normal),
		},
		"rulings": plugins.Option{
			Type:         plugins.OptionTypeBool,
			Description:  "add the rulings to each card description",
			DefaultValue: false,
		},
		"tokens": plugins.Option{
			Type:         plugins.OptionTypeBool,
			Description:  "generate a separate token deck",
			DefaultValue: true,
		},
		"detailed_description": plugins.Option{
			Type:         plugins.OptionTypeBool,
			Description:  "show all card info in the description of the card",
			DefaultValue: false,
		},
	}
}

func (p magicPlugin) URLHandlers() []plugins.URLHandler {
	return []plugins.URLHandler{
		{
			BasePath: "https://scryfall.com",
			Regex:    regexp.MustCompile(`^https://scryfall\.com/@.+/decks/`),
			Handler: func(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
				parsedURL, err := url.Parse(baseURL)
				if err != nil {
					return nil, err
				}

				uuid := path.Base(parsedURL.Path)

				return handleLink(
					baseURL,
					`//h1[contains(@class,'deck-details-title')]`,
					"https://api.scryfall.com/decks/"+uuid+"/export/text",
					options,
				)
			},
		},
		{
			BasePath: "https://deckstats.net",
			Regex:    regexp.MustCompile(`^https://deckstats\.net/decks/`),
			Handler: func(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
				fileURL, err := url.Parse(baseURL)
				if err != nil {
					return nil, err
				}
				q := fileURL.Query()
				q.Set("include_comments", "1")
				q.Set("export_mtgarena", "1")
				fileURL.RawQuery = q.Encode()

				return handleLink(
					baseURL,
					`//h2[@id='subtitle']`,
					fileURL.String(),
					options,
				)
			},
		},
		{
			BasePath: "https://tappedout.net",
			Regex:    regexp.MustCompile(`^https?://tappedout\.net/(?:mtg-decks|mtg-cube-drafts)/`),
			Handler: func(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
				fileURL, err := url.Parse(baseURL)
				if err != nil {
					return nil, err
				}
				q := fileURL.Query()
				q.Set("fmt", "csv")
				fileURL.RawQuery = q.Encode()

				var titleXPath string
				if strings.Contains(baseURL, "mtg-cube-draft") {
					// Cubes
					titleXPath = `//div[contains(@class,'jumbotron')]//div[contains(@class,'row')]/h1`
				} else {
					// Decks
					titleXPath = `//div[contains(@class,'well')]/h2`
				}

				return handleCSVLink(
					baseURL,
					titleXPath,
					fileURL.String(),
					options,
				)
			},
		},
		{
			BasePath: "https://deckbox.org",
			Regex:    regexp.MustCompile(`^https://deckbox\.org/sets/`),
			Handler: func(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
				var fileURL string

				if strings.HasSuffix(baseURL, "/") {
					fileURL = baseURL + "export"
				} else {
					fileURL = baseURL + "/export"
				}

				return handleHTMLLink(
					baseURL,
					`//div[contains(@class,'section_title')][1]/span[1]`,
					fileURL,
					options,
				)
			},
		},
		{
			BasePath: "https://www.mtggoldfish.com",
			Regex:    regexp.MustCompile(`^https://www\.mtggoldfish\.com/deck/`),
			Handler: func(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
				return handleLinkWithDownloadLink(
					baseURL,
					`//h1[contains(@class,'title')]/text()`,
					`//a[contains(text(),'Download')]/@href`,
					"https://www.mtggoldfish.com",
					options,
				)
			},
		},
		{
			BasePath: "https://www.moxfield.com",
			Regex:    regexp.MustCompile(`^https://www\.moxfield\.com/decks/`),
			Handler:  handleMoxfieldLink,
		},
		{
			BasePath: "https://manastack.com",
			Regex:    regexp.MustCompile(`^https://manastack\.com/deck/`),
			Handler:  handleManaStackLink,
		},
		{
			BasePath: "https://archidekt.com",
			Regex:    regexp.MustCompile(`^https://(?:www\.)?archidekt\.com/decks/\d+`),
			Handler:  handleArchidektLink,
		},
		{
			BasePath: "https://aetherhub.com",
			Regex:    regexp.MustCompile(`^https://aetherhub\.com/.*Deck/`),
			Handler:  handleAetherHubLink,
		},
		{
			BasePath: "https://www.frogtown.me",
			Regex:    regexp.MustCompile(`^https://www\.frogtown\.me/deckViewer/`),
			Handler:  handleFrogtownLink,
		},
		{
			BasePath: "https://www.cubetutor.com",
			Regex:    regexp.MustCompile(`^https://www\.cubetutor\.com/(?:viewcube|cubedeck)/`),
			Handler: func(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
				var (
					deckName     string
					cardSetXPath string
					cardsXPath   string
				)

				titleXPath := `//div[@id='main']//h1`

				log.Infof("Checking %s", baseURL)
				doc, err := htmlquery.LoadURL(baseURL)
				if err != nil {
					return nil, fmt.Errorf("couldn't query %s: %w", baseURL, err)
				}

				// Find the title
				title := htmlquery.FindOne(doc, titleXPath)
				if title == nil {
					return nil, fmt.Errorf("no title found in %s (XPath: %s)", baseURL, titleXPath)
				}
				titleText := htmlquery.InnerText(title)

				if strings.Contains(baseURL, "viewcube") {
					// Cubes
					cardSetXPath = `//div[@id='main']`
					cardsXPath = `//a[contains(@class,'cardPreview')]`

					// Parse the name
					split := strings.Split(titleText, "(")
					deckName = strings.TrimSpace(split[0])

					log.Infof("Found title: %s", deckName)
				} else {
					// Decks
					cardSetXPath = `//div[contains(@class,'cardset')]`
					cardsXPath = `//div[contains(@class,'card')]//img/@src`

					// Parse the name
					split := strings.Split(titleText, " by ")
					deckName = strings.TrimSpace(split[0])
					author := strings.TrimSpace(split[1])

					log.Infof("Found title: %s (created by %s)", deckName, author)
				}

				return handleCubeTutorLink(doc, baseURL, deckName, cardSetXPath, cardsXPath, options)
			},
		},
		{
			BasePath: "https://cubecobra.com",
			Regex:    regexp.MustCompile(`^https://cubecobra\.com/cube/(?:overview|list|playtest|analysis|blog)/[^\/]+$`),
			Handler:  handleCubeCobraLink,
		},
		{
			BasePath: "https://mtg.wtf/deck",
			Regex:    regexp.MustCompile(`^https://mtg\.wtf/deck/`),
			Handler: func(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
				var fileURL string

				if strings.HasSuffix(baseURL, "/") {
					fileURL = baseURL + "download"
				} else {
					fileURL = baseURL + "/download"
				}

				return handleHTMLLink(
					baseURL,
					`//header/h4/text()`,
					fileURL,
					options,
				)
			},
		},
	}
}

func (p magicPlugin) FileExtHandlers() map[string]plugins.FileHandler {
	return map[string]plugins.FileHandler{
		".dec": fromDeckFile,
		".cod": fromCockatriceDeckFile,
	}
}

func (p magicPlugin) DeckTypeHandlers() map[string]plugins.DeckType {
	return map[string]plugins.DeckType{
		"Cockatrice": {
			FileHandler: fromCockatriceDeckFile,
			Example: `<?xml version="1.0" encoding="UTF-8"?>
<cockatrice_deck version="1">
    <deckname></deckname>
    <comments></comments>
    <zone name="main">
        <card number="4" name="&quot;Ach! Hans, Run!&quot;"/>
        <card number="1" name="1996 World Champion"/>
        <card number="1" name="Aboshan's Desire"/>
        <card number="10" name="Swamp"/>
    </zone>
    <zone name="side">
        <card number="4" name="Abbey Gargoyles"/>
        <card number="1" name="Chalice of Life"/>
    </zone>
</cockatrice_deck>`,
		},
	}
}

func (p magicPlugin) GenericFileHandler() plugins.DeckType {
	return plugins.DeckType{
		FileHandler: fromDeckFile,
		Example: `1 Jace, the Mind Sculptor
12 Swamp (2XN) 377

Sideboard
1 Lurrus of the Dream-Den

Maybeboard
1 Yorion, Sky Nomad`,
	}
}

func (p magicPlugin) AvailableBacks() map[string]plugins.Back {
	return map[string]plugins.Back{
		plugins.DefaultBackKey: {
			URL:         defaultBackURL,
			Description: "standard paper card back",
		},
		"planechase": {
			URL:         planechaseBackURL,
			Description: "Planechase Plane / Phenomenon card back",
		},
		"archenemy": {
			URL:         archenemyBackURL,
			Description: "Archenemy Scheme card back",
		},
		"m_filler": {
			URL:         mFillerBackURL,
			Description: "filler card back with a white M in the middle",
		},
	}
}

// MagicPlugin is the exported plugin for this package
var MagicPlugin = magicPlugin{
	id:   "mtg",
	name: "Magic",
}
