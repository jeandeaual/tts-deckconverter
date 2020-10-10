package mtg

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	scryfall "github.com/BlueMonday/go-scryfall"
	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"golang.org/x/net/html"

	"github.com/jeandeaual/tts-deckconverter/log"
	"github.com/jeandeaual/tts-deckconverter/plugins"
)

const (
	defaultBackURL    = "http://cloud-3.steamusercontent.com/ugc/998016607072060763/7AFEF2CE9E7A7DB735C93CF33CC4C378CBF4B20D/"
	planechaseBackURL = "http://cloud-3.steamusercontent.com/ugc/998016607072060000/1713AE8643632456D06F1BBA962C5514DD8CCC76/"
	archenemyBackURL  = "http://cloud-3.steamusercontent.com/ugc/998016607072055936/0598975AB8EC26E8956D84F9EC73BBE5754E6C80/"
	// M filler card back
	// See http://www.magiclibrarities.net/348-rarities-filler-cards-english-cards-fillers.html
	mFillerBackURL = "http://cloud-3.steamusercontent.com/ugc/998016607072059554/6BF846C387B045FF524AE42758F6962FE3774CDB/"
)

var cardLineRegexps = []*regexp.Regexp{
	// Magic Arena format
	regexp.MustCompile(`^\s*(?P<Count>\d+)\s+(?P<Name>.+)\s+\((?P<Set>[A-Z0-9_]+)\)(\s+(?P<NumberInSet>[\d]+[abs★]*))?$`),
	// Magic Workstation format
	regexp.MustCompile(`^(?P<Sideboard>SB:)?\s*(?P<Count>\d+)\s+\[(?P<Set>[A-Z0-9_]+)\]\s+(?P<Name>.+)$`),
	// Standard format (MTGO, etc.)
	regexp.MustCompile(`^(?P<Sideboard>SB:)?\s*(?P<Count>\d+)x?\s+(?P<Name>[^#]+)(\s+#(?P<Comment>.*))?$`),
}

// DeckType is the type of a parsed deck.
type DeckType int

const (
	// Main deck
	Main DeckType = iota
	// Sideboard deck
	Sideboard
	// Maybeboard cards
	Maybeboard
)

var (
	sets      map[string]scryfall.Set
	setsMutex sync.Mutex
)

func getSets(ctx context.Context, client *scryfall.Client) (map[string]scryfall.Set, error) {
	setsMutex.Lock()
	defer setsMutex.Unlock()

	if sets == nil {
		setList, err := listSets(ctx, client)
		if err != nil {
			return nil, err
		}
		sets = make(map[string]scryfall.Set)
		for _, set := range setList {
			sets[set.Code] = set
		}
	}

	return sets, nil
}

// CardInfo contains the name of a card and its set.
type CardInfo struct {
	// Name of the card.
	Name string
	// Set of the card.
	Set *string
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
func (c *CardNames) Insert(name string, set *string) {
	c.InsertCount(name, set, 1)
}

// InsertCount inserts several new cards in a CardNames struct.
func (c *CardNames) InsertCount(name string, set *string, count int) {
	_, found := c.Counts[name]
	if !found {
		c.Names = append(c.Names, CardInfo{
			Name: name,
			Set:  set,
		})
		c.Counts[name] = count
	} else {
		c.Counts[name] = c.Counts[name] + count
	}
}

// String representation of a CardNames struct.
func (c *CardNames) String() string {
	var sb strings.Builder

	for _, cardInfo := range c.Names {
		count := c.Counts[cardInfo.Name]
		sb.WriteString(strconv.Itoa(count))
		sb.WriteString(" ")
		sb.WriteString(cardInfo.Name)
		sb.WriteString("\n")
	}

	return sb.String()
}

func getImageURL(
	uris *scryfall.ImageURIs,
	highResAvailable bool,
	imageQuality string,
) string {
	if uris == nil {
		log.Warn("No image data available")
		return ""
	}

	var imageURL string

	switch imageQuality {
	case string(small):
		imageURL = uris.Small
	case string(normal):
		imageURL = uris.Normal
	case string(large):
		if highResAvailable {
			imageURL = uris.Large
		} else {
			log.Warn("High-resolution image not available, using normal quality instead of large")
			imageURL = uris.Normal
		}
	case string(png):
		if highResAvailable {
			imageURL = uris.PNG
		} else {
			log.Warn("High-resolution image not available, using normal quality instead of png")
			imageURL = uris.Normal
		}
	}

	return imageURL
}

func checkRulings(ctx context.Context, client *scryfall.Client, cardID string, options map[string]interface{}) ([]scryfall.Ruling, error) {
	var (
		rulings []scryfall.Ruling
		err     error
	)

	// Check the options to see if we want the rulings
	if showRulings, found := options["rulings"]; found && showRulings.(bool) {
		log.Debugf("Querying rulings for card ID %s", cardID)
		rulings, err = getRulings(ctx, client, cardID)
	}

	return rulings, err
}

func parseRelatedTokenIDs(card scryfall.Card) []string {
	tokenIDs := make([]string, 0, len(card.AllParts))

	for _, part := range card.AllParts {
		if part.Component == scryfall.ComponentToken ||
			(part.Component == scryfall.ComponentComboPiece && strings.HasPrefix(part.TypeLine, "Emblem")) {
			uriParts := strings.Split(part.URI, "/")
			tokenIDs = append(tokenIDs, uriParts[len(uriParts)-1])
		}
	}

	return tokenIDs
}

func buildMeldCard(
	ctx context.Context,
	client *scryfall.Client,
	card scryfall.Card,
	rulings []scryfall.Ruling,
	imageQuality string,
	count int,
	deck *plugins.Deck,
) (plugins.CardInfo, error) {
	// Meld card
	// Find the URL of the meld_result
	if len(card.AllParts) == 0 {
		return plugins.CardInfo{}, fmt.Errorf("no meld parts found for card %s", card.Name)
	}

	var meldResultURI string
	for _, part := range card.AllParts {
		if part.Component == scryfall.ComponentMeldResult {
			meldResultURI = part.URI
			break
		}
	}

	if len(meldResultURI) == 0 {
		return plugins.CardInfo{}, fmt.Errorf("no meld result found for card %s", card.Name)
	}

	uriParts := strings.Split(meldResultURI, "/")
	meldResultID := uriParts[len(uriParts)-1]

	log.Debugf("Querying meld result (card ID %s)", meldResultID)

	meldResult, err := getCard(ctx, client, meldResultID)
	if err != nil {
		return plugins.CardInfo{}, fmt.Errorf("Scryfall client error: %v (card ID %s)", err, meldResultID)
	}

	imageURL := getImageURL(card.ImageURIs, card.HighresImage, imageQuality)
	meldResultImageURL := getImageURL(meldResult.ImageURIs, meldResult.HighresImage, imageQuality)

	if len(deck.ThumbnailURL) == 0 {
		deck.ThumbnailURL = meldResult.ImageURIs.PNG
	}

	return plugins.CardInfo{
		Name:        buildCardName(card),
		Description: buildCardDescription(card, rulings),
		ImageURL:    imageURL,
		Count:       count,
		AlternativeState: &plugins.CardInfo{
			Name:        meldResult.Name,
			Description: buildCardDescription(meldResult, rulings),
			ImageURL:    meldResultImageURL,
			Oversized:   true,
		},
	}, nil
}

func buildDoubleFacedCard(
	card scryfall.Card,
	rulings []scryfall.Ruling,
	imageQuality string,
	count int,
	deck *plugins.Deck,
) (plugins.CardInfo, error) {
	if len(card.CardFaces) != 2 {
		return plugins.CardInfo{}, fmt.Errorf("invalid number of card faces: %d", len(card.CardFaces))
	}

	front := card.CardFaces[0]
	back := card.CardFaces[1]

	frontImageURL := getImageURL(&front.ImageURIs, card.HighresImage, imageQuality)
	backImageURL := getImageURL(&back.ImageURIs, card.HighresImage, imageQuality)

	return plugins.CardInfo{
		Name:        buildCardFaceName(front.Name, card.CMC, front.TypeLine),
		Description: buildCardFaceDescription(front, rulings),
		ImageURL:    frontImageURL,
		Count:       count,
		AlternativeState: &plugins.CardInfo{
			Name:        buildCardFaceName(back.Name, card.CMC, back.TypeLine),
			Description: buildCardFaceDescription(back, rulings),
			ImageURL:    backImageURL,
		},
	}, nil
}

func buildSingleFacedCard(
	card scryfall.Card,
	rulings []scryfall.Ruling,
	imageQuality string,
	count int,
	deck *plugins.Deck,
) (plugins.CardInfo, error) {
	var (
		name        string
		description string
	)

	if len(card.CardFaces) > 1 {
		// For flip, split and adventure layouts
		name = buildCardFacesName(card)
		description = buildCardFacesDescription(card.CardFaces, rulings)
	} else {
		// For standard cards
		name = buildCardName(card)
		description = buildCardDescription(card, rulings)
	}

	imageURL := getImageURL(card.ImageURIs, card.HighresImage, imageQuality)

	if len(deck.ThumbnailURL) == 0 {
		deck.ThumbnailURL = card.ImageURIs.PNG
	}

	return plugins.CardInfo{
		Name:        name,
		Description: description,
		ImageURL:    imageURL,
		Count:       count,
		Oversized:   card.Oversized,
	}, nil
}

func cardNamesToDeck(cards *CardNames, name string, options map[string]interface{}) (*plugins.Deck, []string, error) {
	ctx := context.Background()
	deck := &plugins.Deck{
		Name:     name,
		BackURL:  MagicPlugin.AvailableBacks()[plugins.DefaultBackKey].URL,
		CardSize: plugins.CardSizeStandard,
		Rounded:  true,
	}
	tokenIDs := []string{}
	client, err := scryfall.NewClient()
	if err != nil {
		return deck, tokenIDs, err
	}

	imageQuality := MagicPlugin.AvailableOptions()["quality"].DefaultValue.(string)
	if quality, found := options["quality"]; found {
		imageQuality = quality.(string)
	}

	for _, cardInfo := range cards.Names {
		count := cards.Counts[cardInfo.Name]

		opts := scryfall.GetCardByNameOptions{}
		if cardInfo.Set != nil {
			sets, err := getSets(ctx, client)
			if err != nil {
				return deck, tokenIDs, err
			}
			setName := strings.ToLower(*cardInfo.Set)
			// Manual fix for some deckstats.net set names which differ from Scryfall set names.
			// See https://deckstats.net/sets/?lng=en and https://scryfall.com/sets
			if setName == "frf_ugin" {
				setName = "ugin"
			} else if setName == "mps_akh" {
				setName = "mp2"
			} else if strings.Contains(setName, "_") {
				setName = strings.Split(setName, "_")[0]
			}
			if _, found := sets[setName]; found {
				opts.Set = setName
			} else {
				for _, set := range sets {
					if set.MTGOCode != nil && *set.MTGOCode == setName {
						opts.Set = set.Code
						break
					}
					if set.ArenaCode != nil && *set.ArenaCode == setName {
						opts.Set = set.Code
						break
					}
				}
				if len(opts.Set) == 0 {
					log.Warnf("Set code \"%s\" not found", *cardInfo.Set)
				}
			}
		}

		log.Debugf("Querying card %s (set: %s)", cardInfo.Name, opts.Set)

		card, err := getCardByName(ctx, client, cardInfo.Name, opts)
		if err != nil {
			log.Errorw(
				"Scryfall client error",
				"error", err,
				"name", cardInfo.Name,
				"options", opts,
			)
			return deck, tokenIDs, err
		}

		log.Debugf("API response: %v", card)

		switch card.Layout {
		case scryfall.LayoutToken, scryfall.LayoutDoubleFacedToken, scryfall.LayoutEmblem:
			log.Debug("Card is a token, skipping for now")
			tokenIDs = append(tokenIDs, card.ID)
			continue
		}

		// Retrieve the related tokens
		tokenIDs = append(tokenIDs, parseRelatedTokenIDs(card)...)

		rulings, err := checkRulings(ctx, client, card.ID, options)
		if err != nil {
			log.Errorw(
				"Scryfall client error",
				"error", err,
				"name", cardInfo.Name,
				"options", opts,
			)
			continue
		}

		var cardInfo plugins.CardInfo

		switch card.Layout {
		case scryfall.LayoutMeld:
			cardInfo, err = buildMeldCard(ctx, client, card, rulings, imageQuality, count, deck)
		case scryfall.LayoutTransform, scryfall.LayoutDoubleSided, scryfall.LayoutModalDFC:
			// For transform and other two-sided cards
			cardInfo, err = buildDoubleFacedCard(card, rulings, imageQuality, count, deck)
		default:
			cardInfo, err = buildSingleFacedCard(card, rulings, imageQuality, count, deck)
		}

		if err != nil {
			log.Warnf("Couldn't add card to deck: %v", err)
			continue
		}

		deck.Cards = append(deck.Cards, cardInfo)

		log.Infof("Retrieved %s", card.Name)
	}

	return deck, tokenIDs, nil
}

func removeDuplicates(s []string) []string {
	seen := make(map[string]struct{}, len(s))
	i := 0
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		s[i] = v
		i++
	}
	return s[:i]
}

func tokenIDsToDeck(tokenIDs []string, name string, options map[string]interface{}) (*plugins.Deck, error) {
	ctx := context.Background()
	deck := &plugins.Deck{
		Name:     name,
		BackURL:  MagicPlugin.AvailableBacks()[plugins.DefaultBackKey].URL,
		CardSize: plugins.CardSizeStandard,
		Rounded:  true,
	}

	client, err := scryfall.NewClient()
	if err != nil {
		return deck, err
	}

	imageQuality := MagicPlugin.AvailableOptions()["quality"].DefaultValue.(string)
	if quality, found := options["quality"]; found {
		imageQuality = quality.(string)
	}

	tokenIDs = removeDuplicates(tokenIDs)

	for _, tokenID := range tokenIDs {
		log.Debugf("Querying token ID %s", tokenID)

		card, err := getCard(ctx, client, tokenID)
		if err != nil {
			log.Errorw(
				"Scryfall client error",
				"error", err,
				"id", tokenID,
			)
			continue
		}

		rulings, err := checkRulings(ctx, client, card.ID, options)
		if err != nil {
			log.Errorw(
				"Scryfall client error",
				"error", err,
				"id", card.ID,
			)
			continue
		}

		var cardInfo plugins.CardInfo

		if card.Layout == scryfall.LayoutDoubleFacedToken {
			cardInfo, err = buildDoubleFacedCard(card, rulings, imageQuality, 1, deck)
		} else {
			cardInfo, err = buildSingleFacedCard(card, rulings, imageQuality, 1, deck)
		}

		if err != nil {
			log.Warnf("Couldn't add token to deck: %v", err)
			continue
		}

		deck.Cards = append(deck.Cards, cardInfo)
	}

	return deck, nil
}

func fromDeckFile(file io.Reader, name string, options map[string]string) ([]*plugins.Deck, error) {
	// Check the options
	validatedOptions, err := MagicPlugin.AvailableOptions().ValidateNormalize(options)
	if err != nil {
		return nil, err
	}

	main, side, maybe, err := parseDeckFile(file)
	if err != nil {
		return nil, err
	}

	var (
		decks    []*plugins.Deck
		tokenIDs []string
	)

	if main != nil {
		mainDeck, mainTokenIDs, err := cardNamesToDeck(main, name, validatedOptions)
		if err != nil {
			return nil, err
		}

		decks = append(decks, mainDeck)
		tokenIDs = append(tokenIDs, mainTokenIDs...)
	}

	if side != nil {
		sideDeck, sideTokenIDs, err := cardNamesToDeck(side, name+" - Sideboard", validatedOptions)
		if err != nil {
			return nil, err
		}

		decks = append(decks, sideDeck)
		tokenIDs = append(tokenIDs, sideTokenIDs...)
	}

	if maybe != nil {
		maybeDeck, maybeTokenIDs, err := cardNamesToDeck(maybe, name+" - Maybeboard", validatedOptions)
		if err != nil {
			return nil, err
		}

		decks = append(decks, maybeDeck)
		tokenIDs = append(tokenIDs, maybeTokenIDs...)
	}

	if generateTokens, found := validatedOptions["tokens"]; (!found || generateTokens.(bool)) && len(tokenIDs) > 0 {
		tokenDeck, err := tokenIDsToDeck(tokenIDs, name+" - Tokens", validatedOptions)
		if err != nil {
			return nil, err
		}

		decks = append(decks, tokenDeck)
	}

	return decks, nil
}

func parseDeckLine(
	line string,
	main *CardNames,
	side *CardNames,
	maybe *CardNames,
	step DeckType,
	sbLineFound bool,
	emptyLineCount int,
) (
	*CardNames,
	*CardNames,
	*CardNames,
	DeckType,
	bool,
	int,
) {
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
		sideboardIdx := plugins.IndexOf("Sideboard", groupNames)
		if sideboardIdx != -1 && len(matches[sideboardIdx]) > 0 && !sbLineFound {
			step = Sideboard
			log.Debug("Switched to sideboard (found line starting with \"SB:\")")

			if side != nil && len(side.Names) > 0 {
				// This is the first line starting with SB:, but we
				// already have cards in the sideboard
				// That means we found an empty line beforehand,
				// assuming this would be the sideboard separator
				main.Names = append(main.Names, side.Names...)
				for name, count := range side.Counts {
					if originalCount, found := main.Counts[name]; found {
						main.Counts[name] = originalCount + count
					} else {
						main.Counts[name] = count
					}
				}
				side = nil
			}

			sbLineFound = true
		}
		var set *string
		setIdx := plugins.IndexOf("Set", groupNames)
		if setIdx != -1 && len(matches[setIdx]) > 0 {
			if matches[setIdx] == "000" {
				// TappedOut sometimes exports decks with an invalid set
				// number ("000")
				// Ignore it
				log.Debugf("Ignoring set ID %s", matches[setIdx])
			} else {
				set = &matches[setIdx]
			}
		}

		count, err := strconv.Atoi(matches[countIdx])
		if err != nil {
			log.Errorf("Error when parsing count: %s", err)
			continue
		}
		name := strings.TrimSpace(matches[nameIdx])

		// Some formats use 3 slashes for split cards
		// Since Scryfall uses 2 slashes, replace them
		if strings.Contains(name, "///") {
			name = strings.Replace(name, "///", "//", 1)
		}

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
			main.InsertCount(name, set, count)
		} else if step == Sideboard {
			if side == nil {
				side = NewCardNames()
			}
			side.InsertCount(name, set, count)
		} else if step == Maybeboard {
			if maybe == nil {
				maybe = NewCardNames()
			}
			maybe.InsertCount(name, set, count)
		} else {
			log.Errorw(
				"Found card info but deck not specified",
				"line", line,
			)
		}

		break
	}

	return main, side, maybe, step, sbLineFound, emptyLineCount
}

func parseDeckFile(file io.Reader) (*CardNames, *CardNames, *CardNames, error) {
	var (
		main  *CardNames
		side  *CardNames
		maybe *CardNames
	)
	step := Main
	scanner := bufio.NewScanner(file)
	sbLineFound := false
	emptyLineCount := 0

	for scanner.Scan() {
		line := scanner.Text()

		if len(line) == 0 {
			// Empty line
			// If we already found several main deck cards (two or less could be the commanders),
			// this empty line means we switched to the sideboard
			if main != nil && len(main.Names) > 2 {
				if step == Main {
					step = Sideboard
					log.Debug("Switched to sideboard (found empty line)")
				}
				emptyLineCount++
			}
			continue
		}

		if strings.HasPrefix(line, "Sideboard") {
			if step == Main {
				step = Sideboard
				log.Debug("Switched to sideboard (found comment)")
			}
			continue
		}

		if strings.HasPrefix(line, "Maybeboard") {
			step = Maybeboard
			log.Debug("Switched to maybeboard (found comment)")
			continue
		}

		if strings.HasPrefix(line, "//") {
			// Comment, ignore
			continue
		}

		main, side, maybe, step, sbLineFound, emptyLineCount = parseDeckLine(
			line,
			main,
			side,
			maybe,
			step,
			sbLineFound,
			emptyLineCount,
		)
	}

	if side != nil && !sbLineFound && emptyLineCount > 1 {
		// Multiple empty lines with no line starting with "SB:", that means
		// there was no sideboard
		main.Names = append(main.Names, side.Names...)
		for name, count := range side.Counts {
			if originalCount, found := main.Counts[name]; found {
				main.Counts[name] = originalCount + count
			} else {
				main.Counts[name] = count
			}
		}
		side = nil
	}

	if main != nil {
		log.Debugf("Main: %d different card(s)\n%v", len(main.Names), main)
	} else {
		log.Debug("Main: 0 cards")
	}
	if side != nil {
		log.Debugf("Sideboard: %d different card(s)\n%v", len(side.Names), side)
	} else {
		log.Debug("Sideboard: 0 cards")
	}
	if maybe != nil {
		log.Debugf("Maybeboard: %d different card(s)\n%v", len(maybe.Names), side)
	} else {
		log.Debug("Maybeboard: 0 cards")
	}

	if err := scanner.Err(); err != nil {
		log.Error(err)
		return main, side, maybe, err
	}

	return main, side, maybe, nil
}

func queryDeckFile(fileURL string, deckName string, options map[string]string) (decks []*plugins.Deck, err error) {
	// Build the request
	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create request for %s: %w", fileURL, err)
	}

	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't query %s: %w", fileURL, err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("couldn't close the response body: %w", cerr)
		}
	}()

	return fromDeckFile(resp.Body, deckName, options)
}

func handleLink(url, titleXPath, fileURL string, options map[string]string) (decks []*plugins.Deck, err error) {
	log.Infof("Checking %s", url)
	doc, err := htmlquery.LoadURL(url)
	if err != nil {
		return nil, fmt.Errorf("couldn't query %s: %w", url, err)
	}

	// Find the title
	title := htmlquery.FindOne(doc, titleXPath)
	if title == nil {
		return nil, fmt.Errorf("no title found in %s (XPath: %s)", url, titleXPath)
	}
	deckName := strings.TrimSpace(htmlquery.InnerText(title))
	log.Infof("Found title: %s", deckName)

	return queryDeckFile(fileURL, deckName, options)
}

// tappedout.net CSV format
func handleCSVLink(url, titleXPath, fileURL string, options map[string]string) (decks []*plugins.Deck, err error) {
	log.Infof("Checking %s", url)
	doc, err := htmlquery.LoadURL(url)
	if err != nil {
		return nil, fmt.Errorf("couldn't query %s: %w", url, err)
	}

	// Find the title
	title := htmlquery.FindOne(doc, titleXPath)
	if title == nil {
		return nil, fmt.Errorf("no title found in %s (XPath: %s)", url, titleXPath)
	}
	deckName := strings.TrimSpace(htmlquery.InnerText(title))
	log.Infof("Found title: %s", deckName)

	// Build the request
	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create request for %s: %w", fileURL, err)
	}

	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't query %s: %w", fileURL, err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("couldn't close the response body: %w", cerr)
		}
	}()

	type Card struct {
		Name     string
		Quantity int
		Printing *string
	}

	// Parse the CSV
	commanders := make([]Card, 0, 2)
	main := make([]Card, 0)
	sideboard := make([]Card, 0)
	maybeboard := make([]Card, 0)

	reader := csv.NewReader(resp.Body)

	// Read the header
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("couldn't parse CSV file %s: %w", fileURL, err)
	}

	for {
		// Format: Board,Qty,Name,Printing,Foil,Alter,Signed,Condition,Language,Commander
		// Note: the Commander field is optional
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("couldn't parse CSV file %s: %w", fileURL, err)
		}
		if len(record) < 9 {
			return nil, fmt.Errorf("invalid CSV file format for %s", fileURL)
		}

		quantity, err := strconv.Atoi(record[1])
		if err != nil {
			return nil, fmt.Errorf("couldn't parse Qty from CSV row %s: %w", record, err)
		}
		name := record[2]
		printing := record[3]
		commander := len(record) > 9 && record[9] == "True"

		// TODO: Get the card in the appropriate language (reader[8])

		card := Card{
			Name:     name,
			Quantity: quantity,
		}
		if len(printing) > 0 {
			card.Printing = &printing
		}

		if commander {
			commanders = append(commanders, card)
			continue
		}

		switch record[0] {
		case "side":
			// Sideboard
			sideboard = append(sideboard, card)
		case "maybe":
			// Maybeboard
			maybeboard = append(maybeboard, card)
		default:
			// Mainboard
			main = append(main, card)
		}
	}

	var sb strings.Builder

	printCards := func(sb *strings.Builder, cards []Card) {
		for _, card := range cards {
			sb.WriteString(strconv.Itoa(card.Quantity))
			sb.WriteString(" ")
			sb.WriteString(card.Name)
			if card.Printing != nil {
				sb.WriteString(" (")
				sb.WriteString(strings.ToUpper(*card.Printing))
				sb.WriteString(")")
			}
			sb.WriteString("\n")
		}
	}
	printCards(&sb, commanders)
	printCards(&sb, main)
	if len(sideboard) > 0 {
		sb.WriteString("Sideboard\n")
	}
	printCards(&sb, sideboard)
	if len(maybeboard) > 0 {
		sb.WriteString("Maybeboard\n")
	}
	printCards(&sb, maybeboard)

	return fromDeckFile(strings.NewReader(sb.String()), deckName, options)
}

// deckbox.org exports it's decks in HTML for some reason
func handleHTMLLink(url, titleXPath, fileURL string, options map[string]string) ([]*plugins.Deck, error) {
	log.Infof("Checking %s", url)
	doc, err := htmlquery.LoadURL(url)
	if err != nil {
		return nil, fmt.Errorf("couldn't query %s: %w", url, err)
	}

	// Find the title
	title := htmlquery.FindOne(doc, titleXPath)
	if title == nil {
		return nil, fmt.Errorf("no title found in %s (XPath: %s)", fileURL, titleXPath)
	}
	name := strings.TrimSpace(htmlquery.InnerText(title))
	log.Infof("Found title: %s", name)

	// Retrieve the file
	htmlFile, err := htmlquery.LoadURL(fileURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't query %s: %w", fileURL, err)
	}
	body := htmlquery.FindOne(htmlFile, `//body`)
	if body == nil {
		return nil, fmt.Errorf("no body found in %s", fileURL)
	}

	var output func(buf *bytes.Buffer, n *html.Node)
	output = func(buf *bytes.Buffer, n *html.Node) {
		switch n.Type {
		case html.TextNode:
			buf.WriteString(strings.TrimSpace(n.Data))
			return
		case html.ElementNode:
			// Convert <br> and <br/> to newlines
			if n.Data == "br" {
				buf.WriteString("\n")
				return
			}
			if n.Data == "p" {
				buf.WriteString("\n")
			}
		case html.CommentNode:
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			output(buf, child)
		}
		if n.Type == html.ElementNode && n.Data == "p" {
			buf.WriteString("\n")
		}
	}

	var buffer bytes.Buffer
	output(&buffer, body)

	log.Debugf("Retrieved deck: %s", buffer.String())

	return fromDeckFile(bytes.NewReader(buffer.Bytes()), name, options)
}

func handleLinkWithDownloadLink(url, titleXPath, fileXPath, baseURL string, options map[string]string) (decks []*plugins.Deck, err error) {
	log.Infof("Checking %s", url)
	doc, err := htmlquery.LoadURL(url)
	if err != nil {
		return nil, fmt.Errorf("couldn't query %s: %w", url, err)
	}

	// Find the title
	title := htmlquery.FindOne(doc, titleXPath)
	if title == nil {
		return nil, fmt.Errorf("no title found in %s (XPath: %s)", url, titleXPath)
	}
	deckName := strings.TrimSpace(htmlquery.InnerText(title))
	log.Infof("Found title: %s", deckName)

	// Find the download URL
	a := htmlquery.FindOne(doc, fileXPath)
	if a == nil {
		return nil, fmt.Errorf("no download link found in %s (XPath: %s)", url, fileXPath)
	}
	fileURL := baseURL + htmlquery.InnerText(a)
	log.Infof("Found file URL: %s", fileURL)

	return queryDeckFile(fileURL, deckName, options)
}

type moxfieldDeck struct {
	Name            string                  `json:"name"`
	MainboardCount  int                     `json:"mainboardCount"`
	Mainboard       map[string]moxfieldCard `json:"mainboard"`
	SideboardCount  int                     `json:"sideboardCount"`
	Sideboard       map[string]moxfieldCard `json:"sideboard"`
	MaybeboardCount int                     `json:"maybeboardCount"`
	Maybeboard      map[string]moxfieldCard `json:"maybeboard"`
	CompanionsCount int                     `json:"companionsCount"`
	Companions      map[string]moxfieldCard `json:"companions"`
	CommandersCount int                     `json:"commandersCount"`
	Commanders      map[string]moxfieldCard `json:"commanders"`
}

type moxfieldCard struct {
	Quantity  int              `json:"quantity"`
	BoardType string           `json:"boardType"`
	IsFoil    bool             `json:"isFoil"`
	IsAlter   bool             `json:"isAlter"`
	CardInfo  moxfieldCardInfo `json:"card"`
}

type moxfieldCardInfo struct {
	ID         string `json:"id"`
	ScryfallID string `json:"scryfall_id"`
	Set        string `json:"set"`
	Name       string `json:"name"`
}

func handleMoxfieldLink(baseURL string, options map[string]string) (decks []*plugins.Deck, err error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	deckID := path.Base(parsedURL.Path)
	deckInfoURL := "https://api.moxfield.com/v2/decks/all/" + deckID

	// Build the request
	req, err := http.NewRequest("GET", deckInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create request for %s: %w", deckInfoURL, err)
	}

	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't query %s: %w", deckInfoURL, err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("couldn't close the response body: %w", cerr)
		}
	}()

	data := moxfieldDeck{}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse response from %s: %w", deckInfoURL, err)
	}
	deckName := data.Name

	var sb strings.Builder

	printCards := func(sb *strings.Builder, cards map[string]moxfieldCard) {
		for name, card := range cards {
			sb.WriteString(strconv.Itoa(card.Quantity))
			sb.WriteString(" ")
			sb.WriteString(name)
			if len(card.CardInfo.Set) > 0 {
				sb.WriteString(" (")
				sb.WriteString(strings.ToUpper(card.CardInfo.Set))
				sb.WriteString(")")
			}
			sb.WriteString("\n")
		}
	}
	printCards(&sb, data.Commanders)
	printCards(&sb, data.Companions)
	printCards(&sb, data.Mainboard)
	if data.SideboardCount > 0 {
		sb.WriteString("Sideboard\n")
	}
	printCards(&sb, data.Sideboard)
	if data.MaybeboardCount > 0 {
		sb.WriteString("Maybeboard\n")
	}
	printCards(&sb, data.Maybeboard)

	return fromDeckFile(strings.NewReader(sb.String()), deckName, options)
}

type manaStackDeckOwner struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type manaStackSet struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type manaStackCardInfo struct {
	Name string       `json:"name"`
	Set  manaStackSet `json:"set"`
}

type manaStackCard struct {
	Card       manaStackCardInfo `json:"card"`
	Commander  bool              `json:"commander"`
	Sideboard  bool              `json:"sideboard"`
	Maybeboard bool              `json:"maybeboard"`
}

type manaStackDeck struct {
	Cards []manaStackCard    `json:"cards"`
	Name  string             `json:"name"`
	Owner manaStackDeckOwner `json:"owner"`
}

func handleManaStackLink(baseURL string, options map[string]string) (decks []*plugins.Deck, err error) {
	log.Infof("Checking %s", baseURL)

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	slug := path.Base(parsedURL.Path)
	deckInfoURL := "https://manastack.com/api/deck?slug=" + slug

	// Build the request
	req, err := http.NewRequest("GET", deckInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create request for %s: %w", deckInfoURL, err)
	}

	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't query %s: %w", deckInfoURL, err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("couldn't close the response body: %w", cerr)
		}
	}()

	data := manaStackDeck{}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse response from %s: %w", deckInfoURL, err)
	}
	deckName := data.Name

	commanders := make([]string, 0, 2)
	main := make([]string, 0, len(data.Cards))
	sideboard := make([]string, 0, len(data.Cards))
	maybeboard := make([]string, 0, len(data.Cards))

	for _, card := range data.Cards {
		if card.Commander {
			commanders = append(commanders, card.Card.Name)
		} else if card.Sideboard {
			sideboard = append(sideboard, card.Card.Name)
		} else if card.Maybeboard {
			maybeboard = append(maybeboard, card.Card.Name)
		} else {
			main = append(main, card.Card.Name)
		}
	}

	var sb strings.Builder

	printCards := func(sb *strings.Builder, cards []string) {
		for _, card := range cards {
			sb.WriteString("1 ")
			sb.WriteString(card)
			sb.WriteString("\n")
		}
	}
	printCards(&sb, commanders)
	printCards(&sb, main)
	if len(sideboard) > 0 {
		sb.WriteString("Sideboard\n")
	}
	printCards(&sb, sideboard)
	if len(maybeboard) > 0 {
		sb.WriteString("Maybeboard\n")
	}
	printCards(&sb, maybeboard)

	return fromDeckFile(strings.NewReader(sb.String()), deckName, options)
}

type archidektOwner struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

type archidektOracleCard struct {
	Name string `json:"name"`
}

type archidektEdition struct {
	Code     string `json:"editioncode"`
	Name     string `json:"editionname"`
	MTGOCode string `json:"mtgoCode"`
}

type archidektCardInfo struct {
	SkryfallID string              `json:"uid"`
	OracleCard archidektOracleCard `json:"oracleCard"`
	Edition    archidektEdition    `json:"edition"`
}

type archidektCard struct {
	Card     archidektCardInfo `json:"card"`
	Quantity int               `json:"quantity"`
	Modifier string            `json:"modifier"`
	Category string            `json:"category"`
	Label    string            `json:"label"`
}

type archidektDeck struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Owner       archidektOwner  `json:"owner"`
	Cards       []archidektCard `json:"cards"`
}

func handleArchidektLink(baseURL string, options map[string]string) (decks []*plugins.Deck, err error) {
	log.Infof("Checking %s", baseURL)

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	id := path.Base(parsedURL.Path)
	deckInfoURL := "https://archidekt.com/api/decks/" + id + "/small/"

	// Build the request
	req, err := http.NewRequest("GET", deckInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create request for %s: %w", deckInfoURL, err)
	}

	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't query %s: %w", deckInfoURL, err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("couldn't close the response body: %w", cerr)
		}
	}()

	data := archidektDeck{}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse response from %s: %w", deckInfoURL, err)
	}
	deckName := data.Name

	commanders := make([]archidektCard, 0, 2)
	main := make([]archidektCard, 0, len(data.Cards))
	sideboard := make([]archidektCard, 0, len(data.Cards))
	maybeboard := make([]archidektCard, 0, len(data.Cards))

	for _, card := range data.Cards {
		switch card.Category {
		case "Commander":
			commanders = append(commanders, card)
		case "Sideboard":
			sideboard = append(sideboard, card)
		case "Maybeboard":
			maybeboard = append(maybeboard, card)
		default:
			main = append(main, card)
		}
	}

	var sb strings.Builder

	printCards := func(sb *strings.Builder, cards []archidektCard) {
		for _, card := range cards {
			sb.WriteString(strconv.Itoa(card.Quantity))
			sb.WriteString(" ")
			sb.WriteString(card.Card.OracleCard.Name)
			sb.WriteString(" (")
			sb.WriteString(strings.ToUpper(card.Card.Edition.Code))
			sb.WriteString(")")
			sb.WriteString("\n")
		}
	}
	printCards(&sb, commanders)
	printCards(&sb, main)
	if len(sideboard) > 0 {
		sb.WriteString("Sideboard\n")
	}
	printCards(&sb, sideboard)
	if len(maybeboard) > 0 {
		sb.WriteString("Maybeboard\n")
	}
	printCards(&sb, maybeboard)

	return fromDeckFile(strings.NewReader(sb.String()), deckName, options)
}

var (
	aetherHubTitleXPath      *xpath.Expr
	aetherHubTitleMetaXPath  *xpath.Expr
	aetherHubCardListsXPath  *xpath.Expr
	aetherHubCardLinkXPath   *xpath.Expr
	aetherHubCommanderXPath  *xpath.Expr
	aetherHubCardNameXPath   *xpath.Expr
	aetherHubCardSetXPath    *xpath.Expr
	aetherHubCardNumberXPath *xpath.Expr
)

func init() {
	aetherHubTitleXPath = xpath.MustCompile(`(//div[contains(@class,'text-center')]/h3)[1]`)
	aetherHubTitleMetaXPath = xpath.MustCompile(`(//h3[contains(@class,'text-center')])[1]`)
	aetherHubCardListsXPath = xpath.MustCompile(`//div[starts-with(@id,'tab_visual')]/div[contains(@class,'card-container')]`)
	aetherHubCardLinkXPath = xpath.MustCompile(`//a[contains(@class,'cardLink')]`)
	aetherHubCommanderXPath = xpath.MustCompile(`//img[contains(@class,'img-commander')]/parent::a`)
	aetherHubCardNameXPath = xpath.MustCompile(`/@data-card-name`)
	aetherHubCardSetXPath = xpath.MustCompile(`/@data-card-set`)
	aetherHubCardNumberXPath = xpath.MustCompile(`/@data-card-number`)
}

func handleAetherHubLink(baseURL string, options map[string]string) (decks []*plugins.Deck, err error) {
	log.Infof("Checking %s", baseURL)
	doc, err := htmlquery.LoadURL(baseURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't query %s: %w", baseURL, err)
	}

	// Find the title
	var titleXPath *xpath.Expr
	if strings.Contains(baseURL, "/Metagame/") {
		titleXPath = aetherHubTitleMetaXPath
	} else {
		titleXPath = aetherHubTitleXPath
	}
	title := htmlquery.QuerySelector(doc, titleXPath)
	if title == nil {
		return nil, fmt.Errorf("no title found in %s (XPath: %s)", baseURL, aetherHubTitleXPath)
	}

	// Find the card list
	cardLists := htmlquery.QuerySelectorAll(doc, aetherHubCardListsXPath)
	if len(cardLists) == 0 {
		return nil, fmt.Errorf("no card list found in %s (XPath: %s)", baseURL, aetherHubCardListsXPath)
	}

	pageTitle := htmlquery.InnerText(title)
	splitName := strings.Split(pageTitle, " - ")
	deckName := strings.TrimSpace(strings.Join(splitName[1:], " - "))

	type Card struct {
		Name   string
		Set    string
		Number string
	}

	getCardData := func(link *html.Node) (string, string, string, error) {
		var (
			name   string
			set    string
			number string
		)

		nameAttr := htmlquery.QuerySelector(link, aetherHubCardNameXPath)
		if nameAttr == nil {
			return name, set, number, fmt.Errorf("no name found in link %v (XPath: %s", link, aetherHubCardNameXPath)
		}

		setAttr := htmlquery.QuerySelector(link, aetherHubCardSetXPath)
		if setAttr == nil {
			return name, set, number, fmt.Errorf("no set found in link %v (XPath: %s", link, aetherHubCardSetXPath)
		}

		numberAttr := htmlquery.QuerySelector(link, aetherHubCardNumberXPath)
		if numberAttr == nil {
			return name, set, number, fmt.Errorf("no number found in link %v (XPath: %s", link, aetherHubCardNumberXPath)
		}

		return strings.TrimSpace(htmlquery.InnerText(nameAttr)),
			strings.TrimSpace(htmlquery.InnerText(setAttr)),
			strings.TrimSpace(htmlquery.InnerText(numberAttr)),
			nil
	}

	commanders := make([]Card, 0, 2)
	main := make([]Card, 0)
	sideboard := make([]Card, 0)
	maybeboard := make([]Card, 0)

	// Find the commander(s)
	commanderLinks := htmlquery.QuerySelectorAll(doc, aetherHubCommanderXPath)
	if len(commanderLinks) > 0 {
		for _, commanderLink := range commanderLinks {
			name, set, number, err := getCardData(commanderLink)
			if err != nil {
				return nil, fmt.Errorf("couldn't parse commander in %s: %v", baseURL, err)
			}
			commanders = append(commanders, Card{
				Name:   name,
				Set:    set,
				Number: number,
			})
		}
	}

	for _, cardList := range cardLists {
		prevSibling := cardList.PrevSibling
		if prevSibling.Type == html.TextNode {
			prevSibling = prevSibling.PrevSibling
		}
		prevSiblingTitle := strings.TrimSpace(htmlquery.InnerText(prevSibling))

		cardLinks := htmlquery.QuerySelectorAll(cardList, aetherHubCardLinkXPath)
		if len(cardLinks) == 0 {
			continue
		}

		var deck *[]Card

		if prevSibling.Type == html.ElementNode && prevSibling.Data == "h5" && strings.HasPrefix(prevSiblingTitle, "Side") {
			deck = &sideboard
		} else if prevSibling.Type == html.ElementNode && prevSibling.Data == "h5" && strings.HasPrefix(prevSiblingTitle, "Maybe") {
			deck = &maybeboard
		} else {
			deck = &main
		}

		for _, cardLink := range cardLinks {
			name, set, number, err := getCardData(cardLink)
			if err != nil {
				continue
			}
			*deck = append(*deck, Card{
				Name:   name,
				Set:    set,
				Number: number,
			})
		}
	}

	var sb strings.Builder

	printCards := func(sb *strings.Builder, cards []Card) {
		for _, card := range cards {
			sb.WriteString("1 ")
			sb.WriteString(card.Name)
			sb.WriteString(" (")
			sb.WriteString(strings.ToUpper(card.Set))
			sb.WriteString(")\n")
		}
	}
	printCards(&sb, commanders)
	printCards(&sb, main)
	if len(sideboard) > 0 {
		sb.WriteString("Sideboard\n")
	}
	printCards(&sb, sideboard)
	if len(maybeboard) > 0 {
		sb.WriteString("Maybeboard\n")
	}
	printCards(&sb, maybeboard)

	return fromDeckFile(strings.NewReader(sb.String()), deckName, options)
}

type frogtownSubsets struct {
	IDToName             map[string]string `json:"IDToName"`
	IDToSetCode          map[string]string `json:"IDToSetCode"`
	IDToCollectorsNumber map[string]string `json:"IDToCollectorsNumber"`
}

type frogtownDeckDetails struct {
	ID        string          `json:"_id"`
	Name      string          `json:"name"`
	OwnerID   string          `json:"ownerID"`
	Mainboard []string        `json:"mainboard"`
	Sideboard []string        `json:"sideboard"`
	Subsets   frogtownSubsets `json:"subsets"`
}

type frogtownData struct {
	DeckDetails frogtownDeckDetails `json:"deckDetails"`
}

func handleFrogtownLink(baseURL string, options map[string]string) (decks []*plugins.Deck, err error) {
	scriptXPath := `//body/script[not(@src)]`

	log.Infof("Checking %s", baseURL)
	doc, err := htmlquery.LoadURL(baseURL)
	if err != nil {
		return nil, fmt.Errorf("couldn't query %s: %w", baseURL, err)
	}

	// Find the script tag
	scriptTags := htmlquery.Find(doc, scriptXPath)
	if scriptTags == nil {
		return nil, fmt.Errorf("no script tag found in %s (XPath: %s)", baseURL, scriptXPath)
	}

	const (
		scriptPrefix = "var includedData = "
		scriptSuffix = ";"
	)
	var jsonData string

	for _, scriptTag := range scriptTags {
		scriptContents := strings.TrimSpace(htmlquery.InnerText(scriptTag))
		if strings.HasPrefix(scriptContents, scriptPrefix) {
			jsonData = strings.TrimSuffix(
				strings.TrimPrefix(
					scriptContents,
					scriptPrefix,
				),
				scriptSuffix,
			)
			break
		}
	}

	if len(jsonData) == 0 {
		return nil, fmt.Errorf("no includedData found in %s", baseURL)
	}

	var data frogtownData

	err = json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse includedData from %s: %w", baseURL, err)
	}

	deckName := data.DeckDetails.Name

	var sb strings.Builder

	printCards := func(sb *strings.Builder, cards []string) {
		for _, card := range cards {
			name, ok := data.DeckDetails.Subsets.IDToName[card]
			if !ok {
				log.Warnf("card ID %s not found in IDToName: %v", card, data.DeckDetails.Subsets.IDToName)
				continue
			}
			sb.WriteString("1 ")
			sb.WriteString(name)
			sb.WriteString("\n")
		}
	}
	printCards(&sb, data.DeckDetails.Mainboard)
	if len(data.DeckDetails.Sideboard) > 0 {
		sb.WriteString("Sideboard\n")
	}
	printCards(&sb, data.DeckDetails.Sideboard)

	return fromDeckFile(strings.NewReader(sb.String()), deckName, options)
}

var cubeTutorSetRegex *regexp.Regexp = regexp.MustCompile(`^set\d_\d+$`)

func handleCubeTutorLink(doc *html.Node, baseURL string, deckName string, cardSetXPath string, cardsXPath string, options map[string]string) (decks []*plugins.Deck, err error) {
	cardSets := htmlquery.Find(doc, cardSetXPath)
	main := make([]string, 0, 560)
	sideboard := make([]string, 0, 30)
	maybeboard := make([]string, 0, 30)

	for i, cardSet := range cardSets {
		cards := htmlquery.Find(cardSet, cardsXPath)

		for _, card := range cards {
			contents := htmlquery.InnerText(card)
			filename := path.Base(contents)
			cardSlug := strings.TrimSuffix(filename, filepath.Ext(filename))
			cardName, err := url.PathUnescape(cardSlug)
			if err != nil {
				log.Warnf("Invalid card slug %s extracted from element \"%s\"", cardSlug, contents)
				continue
			}

			// Fix for land names
			if strings.HasSuffix(cardName, "1") {
				cardName = cardName[:len(cardName)-1]
			}
			if strings.HasSuffix(cardName, "-full") {
				cardName = cardName[:len(cardName)-5]
			}

			if len(cubeTutorSetRegex.FindString(cardName)) > 0 {
				log.Warnf("Invalid card name: %s", cardName)
				continue
			}

			switch i {
			case 0:
				main = append(main, cardName)
			case 1:
				sideboard = append(sideboard, cardName)
			default:
				maybeboard = append(maybeboard, cardName)
			}
		}
	}

	var sb strings.Builder

	printCards := func(sb *strings.Builder, cards []string) {
		for _, card := range cards {
			sb.WriteString("1 ")
			sb.WriteString(card)
			sb.WriteString("\n")
		}
	}
	printCards(&sb, main)
	if len(sideboard) > 0 {
		sb.WriteString("Sideboard\n")
	}
	printCards(&sb, sideboard)
	if len(maybeboard) > 0 {
		sb.WriteString("Maybeboard\n")
	}
	printCards(&sb, maybeboard)

	return fromDeckFile(strings.NewReader(sb.String()), deckName, options)
}

func handleCubeCobraLink(baseURL string, options map[string]string) (decks []*plugins.Deck, err error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	slug := path.Base(parsedURL.Path)
	titleXPath := `//title`
	fileURL := "https://cubecobra.com/cube/download/mtgo/" + slug

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
	deckName := strings.TrimSpace(strings.Split(titleText, "-")[0])

	log.Infof("Found title: %s", deckName)

	return queryDeckFile(fileURL, deckName, options)
}
