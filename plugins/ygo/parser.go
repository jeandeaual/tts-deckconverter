package ygo

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"golang.org/x/net/html"

	"github.com/jeandeaual/tts-deckconverter/log"
	"github.com/jeandeaual/tts-deckconverter/plugins"
	"github.com/jeandeaual/tts-deckconverter/plugins/ygo/api"
)

const (
	// Credit to https://www.deviantart.com/holycrapwhitedragon/art/Yu-Gi-Oh-Back-Card-Template-695173962
	defaultBackURL           = "http://cloud-3.steamusercontent.com/ugc/998016607072069584/863E293843E7DB475380CA7D024416AA684C6167/"
	tcgBackURL               = "http://cloud-3.steamusercontent.com/ugc/998016607077817519/C47C4D4C243E14917FBA0CF8A396E56662AB3E0A/"
	ocgBackURL               = "http://cloud-3.steamusercontent.com/ugc/998016607077818666/9AADE856EC9E2AE6BF82557A2FA257E5F6967EC9/"
	animeBackURL             = "http://cloud-3.steamusercontent.com/ugc/998016607072063005/EA4E58868A0DB8A94A1243E61434089CF319F37D/"
	ygoproDeckFileXPath      = `//td[contains(text(),'Deck File')]/a/@href`
	ygoproDeckTitleXPath     = `//h1[@class="entry-title"]`
	yugiohTopDecksFileXPath  = `//a[contains(b/text(),'YGOPro')]/@href`
	yugiohTopDecksTitleXPath = `/html/body/div[3]/div[2]/div[1]/div[1]/div/h3/b`
)

var cardLineRegexps = []*regexp.Regexp{
	regexp.MustCompile(`^(?P<Count>[1-4])[xX]?\s+(?P<Name>.+)$`),
	regexp.MustCompile(`^(?P<Name>.+)\s+[xX]?(?P<Count>[1-4])$`),
}
var mainRegex = regexp.MustCompile(`^Main:?$`)
var extraRegex = regexp.MustCompile(`^Extra:?$`)
var sideRegex = regexp.MustCompile(`^Side:?$`)

// DeckType is the type of a parsed deck.
type DeckType int

const (
	// None means the deck type being parsed hasn't been determined yet
	None DeckType = iota
	// Main deck
	Main
	// Extra deck
	Extra
	// Side deck
	Side
)

// CardIDs contains the card IDs and their count.
type CardIDs struct {
	// IDs are the card IDs.
	IDs []int64
	// Counts is a map of card ID to count (number of cards).
	Counts map[int64]int
}

// NewCardIDs creates a new CardIDs struct.
func NewCardIDs() *CardIDs {
	counts := make(map[int64]int)
	return &CardIDs{Counts: counts}
}

// Insert a new card in a CardIDs struct.
func (c *CardIDs) Insert(id int64) {
	c.InsertCount(id, 1)
}

// InsertCount inserts several new cards in a CardIDs struct.
func (c *CardIDs) InsertCount(id int64, count int) {
	_, found := c.Counts[id]
	if !found {
		c.IDs = append(c.IDs, id)
		c.Counts[id] = count
	} else {
		c.Counts[id] = c.Counts[id] + count
	}
}

// Count return the number of cards for a given ID.
func (c *CardIDs) Count(id int64) int {
	return c.Counts[id]
}

// String representation of a CardIDs struct.
func (c *CardIDs) String() string {
	var sb strings.Builder

	for _, id := range c.IDs {
		count := c.Count(id)
		sb.WriteString(strconv.Itoa(count))
		sb.WriteString("x ")
		sb.WriteString(strconv.FormatInt(id, 10))
		sb.WriteString("\n")
	}

	return sb.String()
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

func cardIDsToDeck(cards *CardIDs, deckName string, format api.Format) (*plugins.Deck, []plugins.CardInfo, error) {
	deck := &plugins.Deck{
		Name:     deckName,
		BackURL:  YGOPlugin.AvailableBacks()[plugins.DefaultBackKey].URL,
		CardSize: plugins.CardSizeSmall,
		Rounded:  false,
	}
	var tokens []plugins.CardInfo

	for _, id := range cards.IDs {
		count := cards.Count(id)

		log.Debugf("Querying card ID %d", id)

		resp, err := queryID(id, format)
		if err != nil {
			return deck, tokens, fmt.Errorf("couldn't query card ID %d (format: %s): %w", id, format, err)
		}

		log.Debugf("API response: %+v", resp)

		if resp.Type == api.TypeToken {
			if len(resp.Images) == 1 {
				tokens = append(tokens, plugins.CardInfo{
					Name:        resp.Name,
					Description: buildDescription(resp),
					ImageURL:    resp.Images[0].URL,
					Count:       count,
				})
			} else {
				// Iterate through each token image
				for i := 0; i < count; i++ {
					tokens = append(tokens, plugins.CardInfo{
						Name:        resp.Name,
						Description: buildDescription(resp),
						ImageURL:    resp.Images[i%len(resp.Images)].URL,
						Count:       1,
					})
				}
			}
		} else {
			deck.Cards = append(deck.Cards, plugins.CardInfo{
				Name:        resp.Name,
				Description: buildDescription(resp),
				ImageURL:    resp.Images[0].URL,
				Count:       count,
			})
		}

		log.Infof("Retrieved %d", id)
	}

	return deck, tokens, nil
}

func cardNamesToDeck(cards *CardNames, deckName string, format api.Format) (*plugins.Deck, []plugins.CardInfo, error) {
	deck := &plugins.Deck{
		Name:     deckName,
		BackURL:  YGOPlugin.AvailableBacks()[plugins.DefaultBackKey].URL,
		CardSize: plugins.CardSizeSmall,
		Rounded:  false,
	}
	var tokens []plugins.CardInfo

	for _, name := range cards.Names {
		count := cards.Count(name)

		log.Debugf("Querying card name %s", name)

		resp, err := queryName(name, format)
		if err != nil {
			return deck, tokens, fmt.Errorf("couldn't query card %s (format: %s): %w", name, format, err)
		}

		log.Debugf("API response: %+v", resp)

		if resp.Type == api.TypeToken {
			if len(resp.Images) == 1 {
				tokens = append(tokens, plugins.CardInfo{
					Name:        resp.Name,
					Description: buildDescription(resp),
					ImageURL:    resp.Images[0].URL,
					Count:       count,
				})
			} else {
				// Iterate through each token image
				for i := 0; i < count; i++ {
					tokens = append(tokens, plugins.CardInfo{
						Name:        resp.Name,
						Description: buildDescription(resp),
						ImageURL:    resp.Images[i%len(resp.Images)].URL,
						Count:       1,
					})
				}
			}
		} else {
			deck.Cards = append(deck.Cards, plugins.CardInfo{
				Name:        resp.Name,
				Description: buildDescription(resp),
				ImageURL:    resp.Images[0].URL,
				Count:       count,
			})
		}

		log.Infof("Retrieved %s", name)
	}

	return deck, tokens, nil
}

func parseYDKFile(file io.Reader) (*CardIDs, *CardIDs, *CardIDs, error) {
	var (
		main  *CardIDs
		extra *CardIDs
		side  *CardIDs
	)
	step := None
	scanner := bufio.NewScanner(file)
	first := true

	for scanner.Scan() {
		line := scanner.Text()

		if first {
			// Remove the UTF-8 BOM, if present
			line = strings.TrimLeft(line, "\uFEFF")
			first = false
		}

		if step != Main && line == "#main" {
			step = Main
			log.Debug("Switched to main")
			continue
		} else if step != Extra && line == "#extra" {
			step = Extra
			log.Debug("Switched to extra")
			continue
		} else if step != Side && line == "!side" {
			step = Side
			log.Debug("Switched to side")
			continue
		}

		if len(line) == 0 {
			// Empty line, ignore
			continue
		}

		if strings.HasPrefix(line, "#") {
			// Comment, ignore
			continue
		}

		if line == "none" {
			continue
		}

		// Try to parse the ID
		id, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			log.Error(err)
			continue
		}

		if step == Main {
			if main == nil {
				main = NewCardIDs()
			}
			main.Insert(id)
		} else if step == Extra {
			if extra == nil {
				extra = NewCardIDs()
			}
			extra.Insert(id)
		} else if step == Side {
			if side == nil {
				side = NewCardIDs()
			}
			side.Insert(id)
		} else {
			log.Errorw(
				"Found card info but deck not specified",
				"line", line,
			)
		}
	}

	if main != nil {
		log.Debugf("Main: %d different card(s)\n%v", len(main.IDs), main)
	} else {
		log.Debug("Main: 0 cards")
	}
	if extra != nil {
		log.Debugf("Extra: %d different card(s)\n%v", len(extra.IDs), extra)
	} else {
		log.Debug("Extra: 0 cards")
	}
	if side != nil {
		log.Debugf("Side: %d different card(s)\n%v", len(side.IDs), side)
	} else {
		log.Debug("Side: 0 cards")
	}

	if err := scanner.Err(); err != nil {
		log.Error(err)
		return main, extra, side, err
	}

	return main, extra, side, nil
}

func parseDeckFile(file io.Reader) (*CardNames, *CardNames, *CardNames, error) {
	var (
		main  *CardNames
		extra *CardNames
		side  *CardNames
	)
	step := Main
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if step != Main && mainRegex.MatchString(line) {
			step = Main
			log.Debug("Switched to main")
			continue
		} else if step != Extra && extraRegex.MatchString(line) {
			step = Extra
			log.Debug("Switched to side")
			continue
		} else if step != Side && sideRegex.MatchString(line) {
			step = Side
			log.Debug("Switched to extra")
			continue
		}

		if len(line) == 0 {
			// Empty line, ignore
			continue
		}

		if strings.HasPrefix(line, "#") {
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
				"step", step,
				"regex", regex,
				"matches", matches,
				"groupNames", groupNames,
			)

			if step == Main {
				if main == nil {
					main = NewCardNames()
				}
				main.InsertCount(name, count)
			} else if step == Extra {
				if extra == nil {
					extra = NewCardNames()
				}
				extra.InsertCount(name, count)
			} else if step == Side {
				if side == nil {
					side = NewCardNames()
				}
				side.InsertCount(name, count)
			} else {
				log.Errorw(
					"Found card info but deck not specified",
					"line", line,
				)
			}

			break
		}
	}

	if main != nil {
		log.Debugf("Main: %d different card(s)\n%v", len(main.Names), main)
	} else {
		log.Debug("Main: 0 cards")
	}
	if extra != nil {
		log.Debugf("Extra: %d different card(s)\n%v", len(extra.Names), extra)
	} else {
		log.Debug("Extra: 0 cards")
	}
	if side != nil {
		log.Debugf("Side: %d different card(s)\n%v", len(side.Names), side)
	} else {
		log.Debug("Side: 0 cards")
	}

	if err := scanner.Err(); err != nil {
		log.Error(err)
		return main, extra, side, err
	}

	return main, extra, side, nil
}

func fromYDKFile(file io.Reader, name string, options map[string]string) ([]*plugins.Deck, error) {
	main, extra, side, err := parseYDKFile(file)
	if err != nil {
		return nil, err
	}

	duelFormat := api.Format(YGOPlugin.AvailableOptions()["format"].DefaultValue.(string))
	if format, found := options["format"]; found {
		duelFormat = api.Format(format)
	}

	var (
		decks  []*plugins.Deck
		tokens []plugins.CardInfo
	)

	if main != nil {
		mainDeck, mainTokens, err := cardIDsToDeck(main, name, duelFormat)
		if err != nil {
			return nil, err
		}

		decks = append(decks, mainDeck)
		tokens = append(tokens, mainTokens...)
	}

	if extra != nil {
		extraDeck, extraTokens, err := cardIDsToDeck(extra, name+" - Extra", duelFormat)
		if err != nil {
			return nil, err
		}

		decks = append(decks, extraDeck)
		tokens = append(tokens, extraTokens...)
	}

	if side != nil {
		sideDeck, sideTokens, err := cardIDsToDeck(side, name+" - Side", duelFormat)
		if err != nil {
			return nil, err
		}

		decks = append(decks, sideDeck)
		tokens = append(tokens, sideTokens...)
	}

	if len(tokens) > 0 {
		decks = append(decks, &plugins.Deck{
			Name:     name + " - Tokens",
			BackURL:  YGOPlugin.AvailableBacks()[plugins.DefaultBackKey].URL,
			CardSize: plugins.CardSizeSmall,
			Rounded:  false,
			Cards:    tokens,
		})
	}

	return decks, nil
}

func fromDeckFile(file io.Reader, name string, options map[string]string) ([]*plugins.Deck, error) {
	main, extra, side, err := parseDeckFile(file)
	if err != nil {
		return nil, err
	}

	duelFormat := api.Format(YGOPlugin.AvailableOptions()["format"].DefaultValue.(string))
	if format, found := options["format"]; found {
		duelFormat = api.Format(format)
	}

	var (
		decks  []*plugins.Deck
		tokens []plugins.CardInfo
	)

	if main != nil {
		mainDeck, mainTokens, err := cardNamesToDeck(main, name, duelFormat)
		if err != nil {
			return nil, err
		}

		decks = append(decks, mainDeck)
		tokens = append(tokens, mainTokens...)
	}

	if extra != nil {
		extraDeck, extraTokens, err := cardNamesToDeck(extra, name+" - Extra", duelFormat)
		if err != nil {
			return nil, err
		}

		decks = append(decks, extraDeck)
		tokens = append(tokens, extraTokens...)
	}

	if side != nil {
		sideDeck, sideTokens, err := cardNamesToDeck(side, name+" - Side", duelFormat)
		if err != nil {
			return nil, err
		}

		decks = append(decks, sideDeck)
		tokens = append(tokens, sideTokens...)
	}

	if len(tokens) > 0 {
		decks = append(decks, &plugins.Deck{
			Name:     name + " - Tokens",
			BackURL:  YGOPlugin.AvailableBacks()[plugins.DefaultBackKey].URL,
			CardSize: plugins.CardSizeSmall,
			Rounded:  false,
			Cards:    tokens,
		})
	}

	return decks, nil
}

func handleLinkWithYDKFile(url string, doc *html.Node, titleXPath, fileXPath, baseURL string, options map[string]string) (decks []*plugins.Deck, err error) {
	log.Infof("Checking %s for YDK file link", url)

	// Find the title
	title := htmlquery.FindOne(doc, titleXPath)
	if title == nil {
		return nil, fmt.Errorf("couldn't retrieve the title from %s (XPath: %s)", url, titleXPath)
	}

	name := strings.TrimSpace(htmlquery.InnerText(title))
	log.Infof("Found title: %s", name)

	// Find the YDK file URL
	a := htmlquery.FindOne(doc, fileXPath)
	if a == nil {
		return nil, fmt.Errorf("couldn't retrieve the YDK URL from %s (XPath: %s)", url, fileXPath)
	}
	ydkURL := baseURL + htmlquery.InnerText(a)
	log.Infof("Found .ydk URL: %s", ydkURL)

	// Build the request
	req, err := http.NewRequest("GET", ydkURL, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create request for %s: %w", ydkURL, err)
	}

	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't query %s: %w", ydkURL, err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("couldn't close the response body: %w", cerr)
		}
	}()

	return fromYDKFile(resp.Body, name, options)
}

var (
	wikiTitleXPath    *xpath.Expr
	wikiPrefixXPath   *xpath.Expr
	wikiCardListXPath *xpath.Expr
	tableRowsXPath    *xpath.Expr
)

func init() {
	wikiTitleXPath = xpath.MustCompile(`//h1[contains(@class,'page-header__title')]/i`)
	wikiPrefixXPath = xpath.MustCompile(`//h3[normalize-space(text())='Prefix(es)']/following-sibling::node()/ul/li[1]`)
	wikiCardListXPath = xpath.MustCompile(`(//table[@id='Top_table'])[1]|(//div[contains(@class,'tabbertab')])[1]/table`)
	tableRowsXPath = xpath.MustCompile(`//tr`)
}

func handleYGOWikiLink(baseURL string, options map[string]string) ([]*plugins.Deck, error) {
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

	// Find the deck prefix
	prefix := htmlquery.QuerySelector(doc, wikiPrefixXPath)
	if title == nil {
		return nil, fmt.Errorf("no prefix found in %s (XPath: %s)", baseURL, wikiPrefixXPath)
	}

	// Find the card list
	table := htmlquery.QuerySelector(doc, wikiCardListXPath)
	if table == nil {
		return nil, fmt.Errorf("no card list found in %s (XPath: %s)", baseURL, wikiCardListXPath)
	}

	deckName := strings.TrimSpace(htmlquery.InnerText(title))

	log.Infof("Found title: %s", deckName)

	if strings.HasPrefix(strings.TrimSpace(htmlquery.InnerText(prefix)), "RD/") {
		options["format"] = string(api.FormatRushDuel)
	}

	rows := htmlquery.QuerySelectorAll(table, tableRowsXPath)
	if rows == nil {
		return nil, fmt.Errorf("no card found in %s (XPath: %s)", baseURL, tableRowsXPath)
	}

	var (
		sb              strings.Builder
		nameCellIdx     int
		quantityCellIdx int
	)

	// Iterate through the table
	for i, row := range rows {
		if i == 0 {
			// Find where the Name and Amount headers are
			headerIdx := 1
			for header := row.FirstChild; header != nil; header = header.NextSibling {
				if header.Type == html.ElementNode && header.Data == "th" {
					switch strings.TrimSpace(htmlquery.InnerText(header)) {
					case "Name", "English name":
						nameCellIdx = headerIdx
					case "Qty":
						quantityCellIdx = headerIdx
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

		cardAmount := "1"
		if quantityCellIdx > 0 {
			cardQuantityXPath := fmt.Sprintf("/td[%d]", quantityCellIdx)
			quantityCell := htmlquery.FindOne(row, cardQuantityXPath)
			if quantityCell == nil {
				return nil, fmt.Errorf("no card quantity found in row number %d of %s (XPath: %s)", i, baseURL, cardQuantityXPath)
			}
			cardAmount = strings.TrimSpace(htmlquery.InnerText(quantityCell))
		}

		sb.WriteString(cardAmount)
		sb.WriteString("x ")
		sb.WriteString(cardName)
		sb.WriteString("\n")
	}

	return fromDeckFile(strings.NewReader(sb.String()), deckName, options)
}
