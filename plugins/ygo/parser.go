package ygo

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"go.uber.org/zap"

	"deckconverter/plugins"
	"deckconverter/plugins/ygo/api"
)

const (
	// Credit to https://www.deviantart.com/holycrapwhitedragon/art/Yu-Gi-Oh-Back-Card-Template-695173962
	defaultBackURL           = "http://cloud-3.steamusercontent.com/ugc/998016607072069584/863E293843E7DB475380CA7D024416AA684C6167/"
	tcgBackURL               = "http://cloud-3.steamusercontent.com/ugc/998016607077817519/C47C4D4C243E14917FBA0CF8A396E56662AB3E0A/"
	ocgBackURL               = "http://cloud-3.steamusercontent.com/ugc/998016607077818666/9AADE856EC9E2AE6BF82557A2FA257E5F6967EC9/"
	animeBackURL             = "http://cloud-3.steamusercontent.com/ugc/998016607072063005/EA4E58868A0DB8A94A1243E61434089CF319F37D/"
	apiCallInterval          = 100 * time.Millisecond
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

type CardIDs struct {
	IDs    []int64
	Counts map[int64]int
}

func NewCardIDs() *CardIDs {
	counts := make(map[int64]int)
	return &CardIDs{Counts: counts}
}

func (c *CardIDs) Insert(id int64) {
	c.InsertCount(id, 1)
}

func (c *CardIDs) InsertCount(id int64, count int) {
	_, found := c.Counts[id]
	if !found {
		c.IDs = append(c.IDs, id)
		c.Counts[id] = count
	} else {
		c.Counts[id] = c.Counts[id] + count
	}
}

func (c *CardIDs) String() string {
	var sb strings.Builder

	for _, id := range c.IDs {
		count := c.Counts[id]
		sb.WriteString(strconv.Itoa(count))
		sb.WriteString("x ")
		sb.WriteString(strconv.FormatInt(id, 10))
		sb.WriteString("\n")
	}

	return sb.String()
}

type CardNames struct {
	Names  []string
	Counts map[string]int
}

func NewCardNames() *CardNames {
	counts := make(map[string]int)
	return &CardNames{Counts: counts}
}

func (c *CardNames) Insert(name string) {
	c.InsertCount(name, 1)
}

func (c *CardNames) InsertCount(name string, count int) {
	_, found := c.Counts[name]
	if !found {
		c.Names = append(c.Names, name)
		c.Counts[name] = count
	} else {
		c.Counts[name] = c.Counts[name] + count
	}
}

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

func parseFile(path string, options map[string]string, log *zap.SugaredLogger) ([]*plugins.Deck, error) {
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

	if ext == ".ydk" {
		return fromYDKFile(file, name, options, log)
	}

	return fromDeckFile(file, name, options, log)
}

func cardIDsToDeck(cards *CardIDs, name string, log *zap.SugaredLogger) (*plugins.Deck, error) {
	deck := &plugins.Deck{
		Name:     name,
		BackURL:  YGOPlugin.AvailableBacks()[plugins.DefaultBackKey].URL,
		CardSize: plugins.CardSizeSmall,
	}

	for _, id := range cards.IDs {
		count := cards.Counts[id]
		resp, err := api.Query(strconv.FormatInt(id, 10), log)
		if err != nil {
			return deck, err
		}

		log.Debugf("API response: %v", resp)

		deck.Cards = append(deck.Cards, plugins.CardInfo{
			Name:        resp.Name,
			Description: buildDescription(resp),
			ImageURL:    resp.ImageURL,
			Count:       count,
		})

		log.Infof("Retrieved %d", id)

		time.Sleep(apiCallInterval)
	}

	return deck, nil
}

func cardNamesToDeck(cards *CardNames, name string, log *zap.SugaredLogger) (*plugins.Deck, error) {
	deck := &plugins.Deck{
		Name:     name,
		BackURL:  YGOPlugin.AvailableBacks()[plugins.DefaultBackKey].URL,
		CardSize: plugins.CardSizeSmall,
	}

	for _, name := range cards.Names {
		count := cards.Counts[name]
		resp, err := api.Query(name, log)
		if err != nil {
			return deck, err
		}

		log.Debugf("API response: %v", resp)

		deck.Cards = append(deck.Cards, plugins.CardInfo{
			Name:        resp.Name,
			Description: buildDescription(resp),
			ImageURL:    resp.ImageURL,
			Count:       count,
		})

		log.Infof("Retrieved %s", name)

		time.Sleep(apiCallInterval)
	}

	return deck, nil
}

func parseYDKFile(file io.Reader, log *zap.SugaredLogger) (*CardIDs, *CardIDs, *CardIDs, error) {
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

func parseDeckFile(file io.Reader, log *zap.SugaredLogger) (*CardNames, *CardNames, *CardNames, error) {
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

func fromYDKFile(file io.Reader, name string, options map[string]string, log *zap.SugaredLogger) ([]*plugins.Deck, error) {
	main, extra, side, err := parseYDKFile(file, log)
	if err != nil {
		return nil, err
	}

	var decks []*plugins.Deck

	if main != nil {
		mainDeck, err := cardIDsToDeck(main, name, log)
		if err != nil {
			return nil, err
		}

		decks = append(decks, mainDeck)
	}

	if extra != nil {
		extraDeck, err := cardIDsToDeck(extra, name+" - Extra", log)
		if err != nil {
			return nil, err
		}

		decks = append(decks, extraDeck)
	}

	if side != nil {
		sideDeck, err := cardIDsToDeck(side, name+" - Side", log)
		if err != nil {
			return nil, err
		}

		decks = append(decks, sideDeck)
	}

	return decks, nil
}

func fromDeckFile(file io.Reader, name string, options map[string]string, log *zap.SugaredLogger) ([]*plugins.Deck, error) {
	main, extra, side, err := parseDeckFile(file, log)
	if err != nil {
		return nil, err
	}

	var decks []*plugins.Deck

	if main != nil {
		mainDeck, err := cardNamesToDeck(main, name, log)
		if err != nil {
			return nil, err
		}

		decks = append(decks, mainDeck)
	}

	if extra != nil {
		extraDeck, err := cardNamesToDeck(extra, name+" - Extra", log)
		if err != nil {
			return nil, err
		}

		decks = append(decks, extraDeck)
	}

	if side != nil {
		sideDeck, err := cardNamesToDeck(side, name+" - Side", log)
		if err != nil {
			return nil, err
		}

		decks = append(decks, sideDeck)
	}

	return decks, nil
}

func handleLinkWithYDKFile(url, titleXPath, fileXPath, baseURL string, options map[string]string, log *zap.SugaredLogger) ([]*plugins.Deck, error) {
	log.Infof("Checking %s for YDK file link", url)
	doc, err := htmlquery.LoadURL(url)
	if err != nil {
		log.Errorf("Couldn't query %s: %s", url, err)
		return nil, err
	}

	// Find the title
	title := htmlquery.FindOne(doc, titleXPath)
	name := strings.TrimSpace(htmlquery.InnerText(title))
	log.Infof("Found title: %s", name)

	// Find the YDK file URL
	a := htmlquery.FindOne(doc, fileXPath)
	ydkURL := baseURL + htmlquery.InnerText(a)
	log.Infof("Found .ydk URL: %s", ydkURL)

	// Build the request
	req, err := http.NewRequest("GET", ydkURL, nil)
	if err != nil {
		log.Errorf("Couldn't create request for %s: %s", ydkURL, err)
		return nil, err
	}

	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("Couldn't query %s: %s", ydkURL, err)
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	return fromYDKFile(resp.Body, name, options, log)
}
