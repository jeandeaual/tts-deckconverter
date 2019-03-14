package magic

import (
	"deckconverter/plugins"
	"encoding/xml"
	"io"
	"io/ioutil"

	"go.uber.org/zap"
)

// CockatriceDeck is the main tag in a Cockatrice deck file (.cod)
type CockatriceDeck struct {
	XMLName  xml.Name         `xml:"cockatrice_deck"`
	Version  int              `xml:"version,attr"`
	Name     string           `xml:"deckname"`
	Comments string           `xml:"comments"`
	Zones    []CockatriceZone `xml:"zone"`
}

// CockatriceZone is a Cockatrice deck zone (usually "main" or "side")
type CockatriceZone struct {
	XMLName xml.Name         `xml:"zone"`
	Name    string           `xml:"name,attr"`
	Cards   []CockatriceCard `xml:"card"`
}

// CockatriceCard represents a specific card in a Cockatrice deck file
type CockatriceCard struct {
	XMLName xml.Name `xml:"card"`
	Number  int      `xml:"number,attr"`
	Name    string   `xml:"name,attr"`
}

func fromCockatriceDeckFile(file io.Reader, name string, options map[string]string, log *zap.SugaredLogger) ([]*plugins.Deck, error) {
	// Check the options
	validatedOptions, err := MagicPlugin.AvailableOptions().ValidateNormalize(options)
	if err != nil {
		return nil, err
	}

	main, side, err := parseCockatriceDeckFile(file, log)
	if err != nil {
		return nil, err
	}

	var decks []*plugins.Deck

	if main != nil {
		mainDeck, err := cardNamesToDeck(main, name, validatedOptions, log)
		if err != nil {
			return nil, err
		}

		decks = append(decks, mainDeck)
	}

	if side != nil {
		sideDeck, err := cardNamesToDeck(side, name+" - Sideboard", validatedOptions, log)
		if err != nil {
			return nil, err
		}

		decks = append(decks, sideDeck)
	}

	return decks, nil
}

func parseCockatriceDeckFile(file io.Reader, log *zap.SugaredLogger) (*CardNames, *CardNames, error) {
	var (
		main *CardNames
		side *CardNames
		deck CockatriceDeck
	)

	// Read the XML file as a byte array.
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return main, side, err
	}

	// Unmarshal the byte array into a struct
	err = xml.Unmarshal(bytes, &deck)
	if err != nil {
		return main, side, err
	}

	for _, zone := range deck.Zones {
		var selected *CardNames

		switch zone.Name {
		case "main":
			main = NewCardNames()
			selected = main
		case "side":
			side = NewCardNames()
			selected = side
		default:
			log.Warnf("Unknown zone found in Cockatrice file: %s", zone.Name)
			continue
		}

		for _, card := range zone.Cards {
			log.Debugw(
				"Found card",
				"name", card.Name,
				"count", card.Number,
				"zone", zone.Name,
			)

			selected.InsertCount(card.Name, nil, card.Number)
		}
	}

	return main, side, nil
}
