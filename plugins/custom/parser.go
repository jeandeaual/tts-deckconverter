package custom

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/jeandeaual/tts-deckconverter/log"
	"github.com/jeandeaual/tts-deckconverter/plugins"
)

var cardLineRegexps = []*regexp.Regexp{
	regexp.MustCompile(`^\s*(?:(?P<Count>\d+)x?\s+)?(?P<Path>.+)\s+\((?P<Name>.+)\)$`),
	regexp.MustCompile(`^\s*(?:(?P<Count>\d+)x?\s+)?(?P<Path>.+)$`),
}

// CardInfo contains a card file path and its name.
type CardInfo struct {
	// Name of the card.
	Path string
	// Set the card belongs to.
	Name *string
}

// CardFiles contains the card file paths and their count.
type CardFiles struct {
	// Cards are the card names.
	Cards []CardInfo
	// Counts is a map of card name to count (number of this card in the deck).
	Counts map[string]int
}

// NewCardNames creates a new CardFiles struct.
func NewCardNames() *CardFiles {
	counts := make(map[string]int)
	return &CardFiles{Counts: counts}
}

// Insert a new card in a CardFiles struct.
func (c *CardFiles) Insert(path string, name *string) {
	c.InsertCount(path, name, 1)
}

// InsertCount inserts several new cards in a CardFiles struct.
func (c *CardFiles) InsertCount(path string, name *string, count int) {
	_, found := c.Counts[path]
	if !found {
		c.Cards = append(c.Cards, CardInfo{
			Path: path,
			Name: name,
		})
		c.Counts[path] = count
	} else {
		c.Counts[path] = c.Counts[path] + count
	}
}

// String representation of a CardFiles struct.
func (c *CardFiles) String() string {
	var sb strings.Builder

	for _, cardInfo := range c.Cards {
		count := c.Counts[cardInfo.Path]
		sb.WriteString(strconv.Itoa(count))
		sb.WriteString(" ")
		sb.WriteString(cardInfo.Path)
		if cardInfo.Name != nil {
			sb.WriteString("(")
			sb.WriteString(*cardInfo.Name)
			sb.WriteString(")")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func cardFilesToDeck(cards *CardFiles, name string, options map[string]interface{}) (*plugins.Deck, error) {
	deck := &plugins.Deck{
		Name:     name,
		CardSize: plugins.CardSizeStandard,
	}

	for _, cardInfo := range cards.Cards {
		card := plugins.CardInfo{
			ImageURL: cardInfo.Path,
			Count:    cards.Counts[cardInfo.Path],
		}
		if cardInfo.Name != nil {
			card.Name = *cardInfo.Name
		}
		deck.Cards = append(deck.Cards, card)
	}

	return deck, nil
}

func fromList(file io.Reader, name string, options map[string]string) ([]*plugins.Deck, error) {
	// Check the options
	validatedOptions, err := CustomPlugin.AvailableOptions().ValidateNormalize(options)
	if err != nil {
		return nil, err
	}

	main, err := parseList(file)
	if err != nil {
		return nil, err
	}

	var decks []*plugins.Deck

	if main != nil {
		deck, err := cardFilesToDeck(main, name, validatedOptions)
		if err != nil {
			return nil, err
		}

		decks = append(decks, deck)
	}

	return decks, nil
}

func parseList(file io.Reader) (*CardFiles, error) {
	var main *CardFiles
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "//") {
			// Comment, ignore
			continue
		}

		// Try to parse the line
		for _, regex := range cardLineRegexps {
			matches := regex.FindStringSubmatch(line)
			if matches == nil {
				continue
			}

			groupNames := regex.SubexpNames()

			count := 1
			countIdx := plugins.IndexOf("Count", groupNames)
			if countIdx != -1 && len(matches[countIdx]) > 0 {
				var err error
				count, err = strconv.Atoi(matches[countIdx])
				if err != nil {
					log.Errorf("Error when parsing count: %s", err)
					continue
				}
			}

			pathIdx := plugins.IndexOf("Path", groupNames)
			if pathIdx == -1 {
				log.Errorf("path not present in regex: %s", regex)
				continue
			}
			path := matches[pathIdx]

			var name *string
			nameIdx := plugins.IndexOf("Name", groupNames)
			if nameIdx != -1 {
				name = &matches[nameIdx]
			}

			log.Debugw(
				"Found card",
				"path", path,
				"count", count,
				"name", name,
				"regex", regex,
				"matches", matches,
				"groupNames", groupNames,
			)

			if main == nil {
				main = NewCardNames()
			}
			main.InsertCount(path, name, count)

			break
		}
	}

	if main != nil {
		log.Debugf("Main: %d different card(s)\n%v", len(main.Cards), main)
	} else {
		log.Debug("Main: 0 cards")
	}

	if err := scanner.Err(); err != nil {
		log.Error(err)
		return main, err
	}

	return main, nil
}
