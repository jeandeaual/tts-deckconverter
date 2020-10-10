package vanguard

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"github.com/jeandeaual/tts-deckconverter/log"
	"github.com/jeandeaual/tts-deckconverter/plugins"
	"golang.org/x/net/html"
)

const (
	defaultBackURL = "http://cloud-3.steamusercontent.com/ugc/998016607077496987/BA283769E35A4D08BAA11EAFD017A6EC91584C87/"
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

// Count return the number of cards for a given name.
func (c *CardNames) Count(name string) int {
	return c.Counts[name]
}

// String representation of a CardNames struct.
func (c *CardNames) String() string {
	var sb strings.Builder

	for _, name := range c.Names {
		count := c.Count(name)
		sb.WriteString(strconv.Itoa(count))
		sb.WriteString(" ")
		sb.WriteString(name)
		sb.WriteString("\n")
	}

	return sb.String()
}

func cardNamesToDeck(cards *CardNames, name string, options map[string]interface{}) (*plugins.Deck, *plugins.Deck, *plugins.Deck, error) {
	deck := &plugins.Deck{
		Name:     name,
		BackURL:  VanguardPlugin.AvailableBacks()[plugins.DefaultBackKey].URL,
		CardSize: plugins.CardSizeSmall,
		Rounded:  true,
	}
	var (
		gdeck  *plugins.Deck
		tokens *plugins.Deck
	)

	cardLanguage := VanguardPlugin.AvailableOptions()["lang"].DefaultValue.(string)
	if option, found := options["lang"]; found {
		cardLanguage = option.(string)
	}

	vanguardFirst := VanguardPlugin.AvailableOptions()["vanguard-first"].DefaultValue.(bool)
	if option, found := options["vanguard-first"]; found {
		vanguardFirst = option.(bool)
	}

	preferPremium := VanguardPlugin.AvailableOptions()["prefer-premium"].DefaultValue.(bool)
	if option, found := options["prefer-premium"]; found {
		preferPremium = option.(bool)
	}

	for _, cardName := range cards.Names {
		count := cards.Count(cardName)

		log.Debugf("Querying card %s (prefer premium: %v)", cardName, preferPremium)

		card, err := getCard(cardName, preferPremium)
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
			if len(card.JapaneseImageURL) > 0 {
				cardInfo.ImageURL = card.JapaneseImageURL
			} else {
				cardInfo.ImageURL = card.EnglishImageURL
			}
		}

		if card.Type != nil && *card.Type == "G Unit" {
			if gdeck == nil {
				gdeck = &plugins.Deck{
					Name:     name + " - G deck",
					BackURL:  VanguardPlugin.AvailableBacks()[plugins.DefaultBackKey].URL,
					CardSize: plugins.CardSizeSmall,
					Rounded:  true,
				}
			}
			gdeck.Cards = append(gdeck.Cards, cardInfo)
		} else if card.Type != nil && strings.HasPrefix(*card.Type, "Token") ||
			card.Effect != nil && strings.Contains(*card.Effect, "This card is a ticket card") {
			if tokens == nil {
				tokens = &plugins.Deck{
					Name:     name + " - Tokens",
					BackURL:  VanguardPlugin.AvailableBacks()[plugins.DefaultBackKey].URL,
					CardSize: plugins.CardSizeSmall,
					Rounded:  true,
				}
			}
			tokens.Cards = append(tokens.Cards, cardInfo)
		} else {
			deck.Cards = append(deck.Cards, cardInfo)
		}
	}

	if vanguardFirst {
		count := len(deck.Cards)
		vanguard := deck.Cards[count-1]
		deck.Cards = append([]plugins.CardInfo{vanguard}, deck.Cards[:count-1]...)
	}

	return deck, gdeck, tokens, nil
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
		deck, gdeck, tokens, err := cardNamesToDeck(main, name, validatedOptions)
		if err != nil {
			return nil, err
		}

		decks = append(decks, deck)
		if gdeck != nil {
			decks = append(decks, gdeck)
		}
		if tokens != nil {
			decks = append(decks, tokens)
		}
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

	switchDeck := func(deckName string) error {
		log.Infof("Found a new deck: %s", deckName)

		// Found a new deck
		if sb.Len() > 0 {
			var parsedDecks []*plugins.Deck
			parsedDecks, err = fromDeckFile(strings.NewReader(sb.String()), currentDeckName, options)
			if err != nil {
				return err
			}
			decks = append(decks, parsedDecks...)

			// Reset the card list
			sb = strings.Builder{}
		}

		currentDeckName = deckName

		return nil
	}

	for child := mainDiv.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "h4" {
			deckAuthor := strings.TrimSpace(htmlquery.InnerText(child))

			err = switchDeck(deckAuthor)
			if err != nil {
				return decks, err
			}
		} else if child.Type == html.ElementNode && child.Data == "div" && len(child.Attr) > 0 {
			if child.Attr[0].Key == "style" {
				deckAuthor := strings.TrimSpace(htmlquery.InnerText(child))

				err = switchDeck(deckAuthor)
				if err != nil {
					return decks, err
				}
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

	switchDeck := func(deckName string) error {
		log.Infof("Found a new deck: %s", deckName)

		// Found a new deck
		if sb.Len() > 0 {
			var parsedDecks []*plugins.Deck
			parsedDecks, err = fromDeckFile(strings.NewReader(sb.String()), currentDeckName, options)
			if err != nil {
				return err
			}
			decks = append(decks, parsedDecks...)

			// Reset the card list
			sb = strings.Builder{}
		}

		currentDeckName = deckName

		return nil
	}

	for child := mainDiv.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "p" && strings.HasPrefix(htmlquery.InnerText(child), "Deck Name: ") {
			deckName := strings.TrimPrefix(
				strings.TrimSpace(htmlquery.InnerText(child)),
				"Deck Name: ",
			)

			err = switchDeck(deckName)
			if err != nil {
				return decks, err
			}
		} else if child.Type == html.ElementNode && child.Data == "div" && len(child.Attr) > 0 {
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

				err = switchDeck(deckName)
				if err != nil {
					return decks, err
				}
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

var (
	wikiTitleXPath    *xpath.Expr
	wikiCardListXPath *xpath.Expr
	tableRowsXPath    *xpath.Expr
	amountRegexp      *regexp.Regexp
)

func init() {
	wikiTitleXPath = xpath.MustCompile(`//h1[contains(@class,'page-header__title')]`)
	wikiCardListXPath = xpath.MustCompile(`//h2/span[@id='Card_List']/parent::node()/following-sibling::table()`)
	tableRowsXPath = xpath.MustCompile(`//tr`)
	amountRegexp = regexp.MustCompile(`(\d)\s*\+\s*(\d)`)
}

func handleCFVWikiLink(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
	// Set the vanguard first option to false, since we don't know where the first vanguard is
	options["vanguard-first"] = "false"

	log.Infof("Checking %s", baseURL)
	doc, err := htmlquery.LoadURL(baseURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't query %s: %w", baseURL, err)
	}

	// Find the title
	title := htmlquery.QuerySelector(doc, wikiTitleXPath)
	if title == nil {
		return nil, fmt.Errorf("no title found in %s (XPath: %s)", baseURL, wikiTitleXPath)
	}

	// Find the card list
	table := htmlquery.QuerySelector(doc, wikiCardListXPath)
	if table == nil {
		return nil, fmt.Errorf("no card list found in %s (XPath: %s)", baseURL, wikiCardListXPath)
	}

	deckName := strings.TrimSpace(htmlquery.InnerText(title))

	if !strings.HasPrefix(deckName, "V ") {
		options["prefer-premium"] = "true"
	}

	log.Infof("Found title: %s", deckName)

	rows := htmlquery.QuerySelectorAll(table, tableRowsXPath)
	if rows == nil {
		return nil, fmt.Errorf("no card found in %s (XPath: %s)", baseURL, tableRowsXPath)
	}

	var (
		sb            strings.Builder
		nameCellIdx   int
		amountCellIdx int
	)

	// Iterate through the table
	for i, row := range rows {
		if i == 0 {
			// Find where the Name and Amount headers are
			headerIdx := 1
			for header := row.FirstChild; header != nil; header = header.NextSibling {
				if header.Type == html.ElementNode && header.Data == "th" {
					switch strings.TrimSpace(htmlquery.InnerText(header)) {
					case "Name":
						nameCellIdx = headerIdx
					case "Amount":
						amountCellIdx = headerIdx
					}
					headerIdx++
				}
			}
			continue
		}

		cardNameXPath := fmt.Sprintf("/td[%d]/a/@title", nameCellIdx)
		nameCell := htmlquery.FindOne(row, cardNameXPath)
		if nameCell == nil {
			return nil, fmt.Errorf("no card name found in row number %d of %s (XPath: %s)", i, baseURL, cardNameXPath)
		}
		cardName := strings.TrimSpace(htmlquery.InnerText(nameCell))

		cardAmountXPath := fmt.Sprintf("/td[%d]", amountCellIdx)
		amountCell := htmlquery.FindOne(row, cardAmountXPath)
		if amountCell == nil {
			return nil, fmt.Errorf("no card amount found in row number %d of %s (XPath: %s)", i, baseURL, cardAmountXPath)
		}
		cardAmount := strings.TrimSpace(htmlquery.InnerText(amountCell))

		count, err := strconv.Atoi(cardAmount)
		if err != nil {
			count = 0
			// Sometimes amount is an expression like "1+3"
			amounts := amountRegexp.FindAllStringSubmatch(cardAmount, -1)
			if len(amounts) == 0 {
				log.Errorf("Error when parsing amount: %s", err)
				continue
			}
			for i := 1; i < len(amounts[0]); i++ {
				amount, cerr := strconv.Atoi(amounts[0][i])
				if cerr != nil {
					log.Errorf("Error when parsing amount: %s", cerr)
					continue
				}
				count += amount
			}
		}

		sb.WriteString(strconv.Itoa(count))
		sb.WriteString("x ")
		sb.WriteString(cardName)
		sb.WriteString("\n")
	}

	return fromDeckFile(strings.NewReader(sb.String()), deckName, options)
}
