package pkm

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	pokemontcgsdk "github.com/PokemonTCG/pokemon-tcg-sdk-go-v2/pkg"

	"github.com/jeandeaual/tts-deckconverter/log"
	"github.com/jeandeaual/tts-deckconverter/plugins"
)

const (
	defaultBackURL     = "http://cloud-3.steamusercontent.com/ugc/998016607072061655/9BE66430CD3C340060773E321DDD5FD86C1F2703/"
	japaneseBackURL    = "http://cloud-3.steamusercontent.com/ugc/998016607072062006/85BAC9FFDBF402428370296B2FA087285A5BAF7D/"
	japaneseOldBackURL = "http://cloud-3.steamusercontent.com/ugc/998016607072062403/76AA6F40903D2CF105B1FD7C43D071F27CB0A354/"
)

var cardLineRegexps = []*regexp.Regexp{
	// PTCGO format
	regexp.MustCompile(`^\s*\**\s*(?P<Count>\d+)\s+(?P<Name>.+)\s+(?P<Set>[A-Za-z0-9_-]+)\s+(?P<NumberInSet>[A-Za-z0-9]+)$`),
}

// CardInfo contains a card name, its set and its number in a set.
type CardInfo struct {
	// Name of the card.
	Name string
	// Set the card belongs to.
	Set string
	// Number of this card in the deck.
	Number string
}

// CardNames contains the card names and their count.
type CardNames struct {
	// Names are the card names.
	Names []CardInfo
	// Counts is a map of card name to count (number of this card in the deck).
	Counts map[string]int
}

// NewCardNames creates a new CardNames struct.
func NewCardNames() *CardNames {
	counts := make(map[string]int)
	return &CardNames{Counts: counts}
}

// Insert a new card in a CardNames struct.
func (c *CardNames) Insert(name string, set string, number string) {
	c.InsertCount(name, set, number, 1)
}

// InsertCount inserts several new cards in a CardNames struct.
func (c *CardNames) InsertCount(name string, set string, number string, count int) {
	idx := name + set
	_, found := c.Counts[idx]
	if !found {
		c.Names = append(c.Names, CardInfo{
			Name:   name,
			Set:    set,
			Number: number,
		})
		c.Counts[idx] = count
	} else {
		c.Counts[idx] = c.Counts[idx] + count
	}
}

// Count return the number of cards for a given name and set.
func (c *CardNames) Count(name string, set string) int {
	idx := name + set
	return c.Counts[idx]
}

// String representation of a CardNames struct.
func (c *CardNames) String() string {
	var sb strings.Builder

	for _, cardInfo := range c.Names {
		count := c.Count(cardInfo.Name, cardInfo.Set)
		sb.WriteString(strconv.Itoa(count))
		sb.WriteString(" ")
		sb.WriteString(cardInfo.Name)
		sb.WriteString(" ")
		sb.WriteString(cardInfo.Set)
		sb.WriteString(" ")
		sb.WriteString(cardInfo.Number)
		sb.WriteString("\n")
	}

	return sb.String()
}

func cardNamesToDeck(cards *CardNames, name string, options map[string]interface{}) (*plugins.Deck, error) {
	deck := &plugins.Deck{
		Name:     name,
		BackURL:  PokemonPlugin.AvailableBacks()[plugins.DefaultBackKey].URL,
		CardSize: plugins.CardSizeStandard,
		Rounded:  true,
	}

	for _, cardInfo := range cards.Names {
		count := cards.Count(cardInfo.Name, cardInfo.Set)

		set, found := getSetCode(cardInfo.Set)
		if !found {
			set = cardInfo.Set
			// Official set names sometimes contain the "a" or "b" suffix
			set = strings.TrimSuffix(set, "a")
			set = strings.TrimSuffix(set, "b")
			_, found = getPTCGOSetCode(set)
			if !found {
				log.Errorf("Invalid set code: %s", cardInfo.Set)
				continue
			}
		}

		log.Debugf("Querying card %s (%s)", cardInfo.Name, set)

		cards, err := getCards(cardInfo.Name, set)
		if err != nil {
			log.Errorw(
				"Pokemon TCG SDK client error",
				"error", err,
				"name", cardInfo.Name,
				"setCode", set,
			)
			continue
		}

		if len(cards) == 0 {
			log.Errorw(
				"No card found",
				"name", cardInfo.Name,
				"setCode", set,
			)
			continue
		}

		log.Debugf("API response (%d card(s)): %v", len(cards), cards)

		var card pokemontcgsdk.PokemonCard

		for _, card = range cards {
			// If we find the exact number, use this card
			// Otherwise, use the last one
			if card.Number == cardInfo.Number {
				break
			}
		}

		deck.Cards = append(deck.Cards, plugins.CardInfo{
			Name:        card.Name,
			Description: buildCardDescription(card),
			ImageURL:    card.Images.Large,
			Count:       count,
		})
	}

	return deck, nil
}

func fromDeckFile(file io.Reader, name string, options map[string]string) ([]*plugins.Deck, error) {
	// Check the options
	validatedOptions, err := PokemonPlugin.AvailableOptions().ValidateNormalize(options)
	if err != nil {
		return nil, err
	}

	main, err := parseDeckFile(file)
	if err != nil {
		return nil, err
	}

	var decks []*plugins.Deck

	if main != nil {
		deck, err := cardNamesToDeck(main, name, validatedOptions)
		if err != nil {
			return nil, err
		}

		decks = append(decks, deck)
	}

	return decks, nil
}

func parseDeckFile(file io.Reader) (*CardNames, error) {
	var main *CardNames
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
			countIdx := plugins.IndexOf("Count", groupNames)
			if countIdx == -1 {
				log.Errorf("Count not present in regex: %s", regex)
				continue
			}
			nameIdx := plugins.IndexOf("Name", groupNames)
			if nameIdx == -1 {
				log.Errorf("Name not present in regex: %s", regex)
				continue
			}
			setIdx := plugins.IndexOf("Set", groupNames)
			if setIdx == -1 {
				log.Errorf("Set not present in regex: %s", regex)
				continue
			}
			numberIdx := plugins.IndexOf("NumberInSet", groupNames)
			if numberIdx == -1 {
				log.Errorf("Number in set not present in regex: %s", regex)
				continue
			}

			count, err := strconv.Atoi(matches[countIdx])
			if err != nil {
				log.Errorf("Error when parsing count: %s", err)
				continue
			}
			name := strings.TrimSpace(matches[nameIdx])
			set := strings.TrimSpace(matches[setIdx])
			number := strings.TrimSpace(matches[numberIdx])

			log.Debugw(
				"Found card",
				"name", name,
				"set", set,
				"number", number,
				"count", count,
				"regex", regex,
				"matches", matches,
				"groupNames", groupNames,
			)

			if main == nil {
				main = NewCardNames()
			}
			main.InsertCount(name, set, number, count)

			break
		}
	}

	if main != nil {
		log.Debugf("Main: %d different card(s)\n%v", len(main.Names), main)
	} else {
		log.Debug("Main: 0 cards")
	}

	if err := scanner.Err(); err != nil {
		log.Error(err)
		return main, err
	}

	return main, nil
}
