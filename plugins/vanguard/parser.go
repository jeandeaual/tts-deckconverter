package vanguard

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"github.com/jeandeaual/tts-deckconverter/log"
	"github.com/jeandeaual/tts-deckconverter/plugins"
	"github.com/jeandeaual/tts-deckconverter/plugins/vanguard/cardfightwiki"
	"golang.org/x/net/html"
)

const (
	defaultBackURL  = "http://cloud-3.steamusercontent.com/ugc/998016607077496987/BA283769E35A4D08BAA11EAFD017A6EC91584C87/"
	apiCallInterval = 100 * time.Millisecond
)

var cardLineRegexps = []*regexp.Regexp{
	regexp.MustCompile(`^-?\s*(?P<Count>\d+)x?\s+(?P<Name>.+)$`),
}

// CardNames contains the card names and their count.
type CardNames struct {
	// Names are the card names.
	Names []string
	// Counts is a map of card name to count (number of this card in the deck).
	Counts map[string]int
}

// NewCardNames creates a new CardNames struct.
func NewCardNames() *CardNames {
	counts := make(map[string]int)
	return &CardNames{Counts: counts}
}

// Insert a new card in a CardNames struct.
func (c *CardNames) Insert(name string) {
	c.InsertCount(name, 1)
}

// InsertCount inserts several new cards in a CardNames struct.
func (c *CardNames) InsertCount(name string, count int) {
	_, found := c.Counts[name]
	if !found {
		c.Names = append(c.Names, name)
		c.Counts[name] = count
	} else {
		c.Counts[name] = c.Counts[name] + count
	}
}

// String representation of a CardNames struct.
func (c *CardNames) String() string {
	var sb strings.Builder

	for _, name := range c.Names {
		count := c.Counts[name]
		sb.WriteString(strconv.Itoa(count))
		sb.WriteString(" ")
		sb.WriteString(name)
		sb.WriteString("\n")
	}

	return sb.String()
}

func cardNamesToDeck(cards *CardNames, name string, options map[string]interface{}) (*plugins.Deck, error) {
	deck := &plugins.Deck{
		Name:     name,
		BackURL:  VanguardPlugin.AvailableBacks()[plugins.DefaultBackKey].URL,
		CardSize: plugins.CardSizeSmall,
	}

	cardLanguage := VanguardPlugin.AvailableOptions()["lang"].DefaultValue.(string)
	if lang, found := options["lang"]; found {
		cardLanguage = lang.(string)
	}

	for _, cardName := range cards.Names {
		count := cards.Counts[cardName]

		card, err := cardfightwiki.GetCard(cardName, false)
		if err != nil {
			log.Errorw(
				"Cardfight!! Vanguard Wiki parsing error",
				"error", err,
				"name", cardName,
			)
			continue
		}

		log.Debugf("Found card: %v", card)

		cardInfo := plugins.CardInfo{
			Description: buildCardDescription(card),
			Count:       count,
		}
		if cardLanguage == "en" {
			cardInfo.Name = card.EnglishName
			cardInfo.ImageURL = card.EnglishImageURL
		} else {
			cardInfo.Name = card.JapaneseName
			cardInfo.ImageURL = card.JapaneseImageURL
		}

		deck.Cards = append(deck.Cards, cardInfo)

		time.Sleep(apiCallInterval)
	}

	return deck, nil
}

func parseFile(path string, options map[string]string) ([]*plugins.Deck, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	ext := filepath.Ext(path)
	name := strings.TrimSuffix(filepath.Base(path), ext)

	log.Debugf("Base file name: %s", name)

	return fromDeckFile(file, name, options)
}

func fromDeckFile(file io.Reader, name string, options map[string]string) ([]*plugins.Deck, error) {
	// Check the options
	validatedOptions, err := VanguardPlugin.AvailableOptions().ValidateNormalize(options)
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

			count, err := strconv.Atoi(matches[countIdx])
			if err != nil {
				log.Errorf("Error when parsing count: %s", err)
				continue
			}
			name := strings.TrimSpace(matches[nameIdx])

			log.Debugw(
				"Found card",
				"name", name,
				"count", count,
				"regex", regex,
				"matches", matches,
				"groupNames", groupNames,
			)

			if main == nil {
				main = NewCardNames()
			}
			main.InsertCount(name, count)

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

var (
	japaneseMainDivXPath  *xpath.Expr
	englishMainDivXPath   *xpath.Expr
	englishDeckTitleXPath *xpath.Expr
	cardListXPath         *xpath.Expr
	cardNameXPath         *xpath.Expr
	cardCountXPath        *xpath.Expr
)

func init() {
	japaneseMainDivXPath = xpath.MustCompile(`//div[contains(@class,'recipe_data') or contains(@class,'contents-box-main')]`)
	englishMainDivXPath = xpath.MustCompile(`//div[contains(@class,'contents-box-main')]`)
	englishDeckTitleXPath = xpath.MustCompile(`/p[starts-with(.,'Deck Name:')]`)
	cardListXPath = xpath.MustCompile(`//a`)
	cardNameXPath = xpath.MustCompile(`/img/@alt`)
	cardCountXPath = xpath.MustCompile(`/span[contains(@class,'num')]`)
}

func handleCFVanguardLink(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
	// Set the card language to "ja" if not set by the user
	if _, found := options["lang"]; !found {
		options["lang"] = "ja"
	}

	log.Infof("Checking %s", baseURL)
	doc, err := htmlquery.LoadURL(baseURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't query %s: %w", baseURL, err)
	}

	decks := make([]*plugins.Deck, 0, 1)

	// Find the main div
	mainDiv := htmlquery.QuerySelector(doc, japaneseMainDivXPath)
	if mainDiv == nil {
		return nil, fmt.Errorf("couldn't find main div in %s (XPath: %s)", baseURL, japaneseMainDivXPath)
	}

	var (
		sb              strings.Builder
		currentDeckName string
	)

	for child := mainDiv.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "h4" {
			deckAuthor := strings.TrimSpace(htmlquery.InnerText(child))

			// Found a new deck
			if sb.Len() > 0 {
				parsedDecks, err := fromDeckFile(strings.NewReader(sb.String()), currentDeckName, options)
				if err != nil {
					return decks, err
				}
				decks = append(decks, parsedDecks...)

				// Reset the card list
				sb = strings.Builder{}
			}

			currentDeckName = deckAuthor
		} else if child.Type == html.ElementNode && child.Data == "div" && len(child.Attr) > 0 {
			if child.Attr[0].Key == "style" {
				deckAuthor := strings.TrimSpace(htmlquery.InnerText(child))

				// Found a new deck
				if sb.Len() > 0 {
					parsedDecks, err := fromDeckFile(strings.NewReader(sb.String()), currentDeckName, options)
					if err != nil {
						return decks, err
					}
					decks = append(decks, parsedDecks...)

					// Reset the card list
					sb = strings.Builder{}
				}

				currentDeckName = deckAuthor
			} else if child.Attr[0].Key == "class" && child.Attr[0].Val == "recipe-single-list" {
				// Look for the card list
				cardLinks := htmlquery.QuerySelectorAll(child, cardListXPath)
				for _, cardLink := range cardLinks {
					nameNode := htmlquery.QuerySelector(cardLink, cardNameXPath)
					if nameNode == nil {
						log.Warnf("no card name found in card link (URL: %s): %s", baseURL, cardLink)
						continue
					}
					name := strings.TrimSpace(htmlquery.InnerText(nameNode))
					countNode := htmlquery.QuerySelector(cardLink, cardCountXPath)
					if countNode == nil {
						log.Warnf("no card count found in card link (URL: %s): %s", baseURL, cardLink)
						continue
					}
					count := strings.TrimSpace(htmlquery.InnerText(countNode))

					sb.WriteString(count)
					sb.WriteString("x ")
					sb.WriteString(name)
					sb.WriteString("\n")
				}
			}
		}
	}

	parsedDecks, err := fromDeckFile(strings.NewReader(sb.String()), currentDeckName, options)
	if err != nil {
		return decks, err
	}
	decks = append(decks, parsedDecks...)

	return decks, nil
}

func handleENCFVanguardLink(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
	// Set the card language to "en" if not set by the user
	if _, found := options["lang"]; !found {
		options["lang"] = "en"
	}

	log.Infof("Checking %s", baseURL)
	doc, err := htmlquery.LoadURL(baseURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't query %s: %w", baseURL, err)
	}

	decks := make([]*plugins.Deck, 0, 1)

	// Find the main div
	mainDiv := htmlquery.QuerySelector(doc, englishMainDivXPath)
	if mainDiv == nil {
		return nil, fmt.Errorf("couldn't find main div in %s (XPath: %s)", baseURL, englishMainDivXPath)
	}

	var (
		sb              strings.Builder
		currentDeckName string
	)

	for child := mainDiv.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "div" && len(child.Attr) > 0 {
			if child.Attr[0].Key == "class" && child.Attr[0].Val == "recipe-single-desc" {
				// Look for the deck title
				titleNode := htmlquery.QuerySelector(child, englishDeckTitleXPath)
				if titleNode == nil {
					continue
				}
				deckName := strings.TrimPrefix(
					strings.TrimSpace(htmlquery.InnerText(titleNode)),
					"Deck Name: ",
				)

				// Found a new deck
				if sb.Len() > 0 {
					parsedDecks, err := fromDeckFile(strings.NewReader(sb.String()), currentDeckName, options)
					if err != nil {
						return decks, err
					}
					decks = append(decks, parsedDecks...)

					// Reset the card list
					sb = strings.Builder{}
				}

				currentDeckName = deckName
			} else if child.Attr[0].Key == "class" && child.Attr[0].Val == "recipe-single-list" {
				// Look for the card list
				cardLinks := htmlquery.QuerySelectorAll(child, cardListXPath)
				for _, cardLink := range cardLinks {
					nameNode := htmlquery.QuerySelector(cardLink, cardNameXPath)
					if nameNode == nil {
						log.Warnf("no card name found in card link (URL: %s): %s", baseURL, cardLink)
						continue
					}
					name := strings.TrimSpace(htmlquery.InnerText(nameNode))
					countNode := htmlquery.QuerySelector(cardLink, cardCountXPath)
					if countNode == nil {
						log.Warnf("no card count found in card link (URL: %s): %s", baseURL, cardLink)
						continue
					}
					count := strings.TrimSpace(htmlquery.InnerText(countNode))

					sb.WriteString(count)
					sb.WriteString("x ")
					sb.WriteString(name)
					sb.WriteString("\n")
				}
			}
		}
	}

	parsedDecks, err := fromDeckFile(strings.NewReader(sb.String()), currentDeckName, options)
	if err != nil {
		return decks, err
	}
	decks = append(decks, parsedDecks...)

	return decks, nil
}
