package cardfightwiki

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/antchfx/xpath"
	"golang.org/x/net/html"

	"github.com/jeandeaual/tts-deckconverter/log"
)

const (
	wikiBaseURL   = "https://cardfight.fandom.com/wiki/"
	wikiSearchURL = wikiBaseURL + "Special:Search"
)

var (
	searchResultLinkXPath *xpath.Expr
	englishImageURLXPath  *xpath.Expr
	japaneseImageURLXPath *xpath.Expr
	defaultImageURLXPath  *xpath.Expr
	englishNameXPath      *xpath.Expr
	kanjiNameXPath        *xpath.Expr
	kanaNameXPath         *xpath.Expr
	gradeSkillXPath       *xpath.Expr
	cardTypeXPath         *xpath.Expr
	powerXPath            *xpath.Expr
	criticalXPath         *xpath.Expr
	shieldXPath           *xpath.Expr
	nationXPath           *xpath.Expr
	clanXPath             *xpath.Expr
	raceXPath             *xpath.Expr
	triggerEffectXPath    *xpath.Expr
	formatsXPath          *xpath.Expr
	flavorXPath           *xpath.Expr
	effectXPath           *xpath.Expr
	hrefXPath             *xpath.Expr
)

func init() {
	searchResultLinkXPath = xpath.MustCompile(`//a[contains(@class,'unified-search__result__title')]`)
	englishImageURLXPath = xpath.MustCompile(`//span[contains(@class,'English')]/a/@href`)
	japaneseImageURLXPath = xpath.MustCompile(`//span[contains(@class,'Japanese')]/a/@href`)
	defaultImageURLXPath = xpath.MustCompile(`//div[contains(@style,'float:left;')]/a/@href`)
	englishNameXPath = xpath.MustCompile(`//td[normalize-space(text())='Name']/following-sibling::node()[2]`)
	kanjiNameXPath = xpath.MustCompile(`//td[normalize-space(text())='Kanji']/following-sibling::node()[2]`)
	kanaNameXPath = xpath.MustCompile(`//td[normalize-space(text())='Kana']/following-sibling::node()[2]`)
	cardTypeXPath = xpath.MustCompile(`//td[normalize-space(text())='Card Type']/following-sibling::node()[2]`)
	gradeSkillXPath = xpath.MustCompile(`//td[normalize-space(text())='Grade / Skill']/following-sibling::node()[2]`)
	powerXPath = xpath.MustCompile(`//td[normalize-space(text())='Power']/following-sibling::node()[2]`)
	criticalXPath = xpath.MustCompile(`//td[normalize-space(text())='Critical']/following-sibling::node()[2]`)
	shieldXPath = xpath.MustCompile(`//td[normalize-space(text())='Shield']/following-sibling::node()[2]`)
	nationXPath = xpath.MustCompile(`//td[normalize-space(text())='Nation']/following-sibling::node()[2]`)
	clanXPath = xpath.MustCompile(`//td[normalize-space(text())='Clan']/following-sibling::node()[2]`)
	raceXPath = xpath.MustCompile(`//td[normalize-space(text())='Race']/following-sibling::node()[2]`)
	triggerEffectXPath = xpath.MustCompile(`//td[normalize-space(text())='Trigger Effect']/following-sibling::node()[2]`)
	formatsXPath = xpath.MustCompile(`//td[normalize-space(text())='Format']/following-sibling::node()[2]`)
	flavorXPath = xpath.MustCompile(`//table[contains(@class,'flavor')]//td`)
	effectXPath = xpath.MustCompile(`//table[contains(@class,'effect')]//td`)
	hrefXPath = xpath.MustCompile(`/@href`)
}

func trimImageURL(imageURL string) string {
	if idx := strings.LastIndex(imageURL, "/revision/latest?cb="); idx != -1 {
		return imageURL[:idx]
	}
	return imageURL
}

func getLinkFromTag(link *html.Node, linkName string, searchURL string, linkNumber int) (string, error) {
	hrefTag := htmlquery.QuerySelector(link, hrefXPath)

	if hrefTag == nil {
		return "", fmt.Errorf("No href tag found in link number %d (%s) of %s", linkNumber+1, linkName, searchURL)
	}

	return htmlquery.InnerText(hrefTag), nil
}

func search(cardName string, preferPremium bool) (string, error) {
	parsedURL, err := url.Parse(wikiSearchURL)
	if err != nil {
		return "", fmt.Errorf("couldn't parse URL %s: %w", wikiSearchURL, err)
	}
	q := parsedURL.Query()
	q.Set("query", "\""+cardName+"\"")
	parsedURL.RawQuery = q.Encode()

	searchURL := parsedURL.String()

	log.Infof("Searching for card %s with %s", cardName, searchURL)

	searchResult, err := htmlquery.LoadURL(searchURL)
	if err != nil {
		return "", fmt.Errorf("couldn't query %s: %w", searchURL, err)
	}

	links := htmlquery.QuerySelectorAll(searchResult, searchResultLinkXPath)
	if links == nil {
		return "", fmt.Errorf("no result link found in %s (XPath: %s)", searchURL, searchResultLinkXPath)
	}

	for i, link := range links {
		linkName := htmlquery.InnerText(link)

		if linkName == cardName+" (V Series)" && !preferPremium {
			href, err := getLinkFromTag(link, linkName, searchURL, i)
			if err != nil {
				log.Warn(err)
				continue
			}

			log.Debugf("Found link %s for card %s", href, cardName)

			return href, nil
		}

		if cardName == linkName {
			href, err := getLinkFromTag(link, linkName, searchURL, i)
			if err != nil {
				log.Warn(err)
				continue
			}

			log.Debugf("Found link %s for card %s", href, cardName)

			return href, nil
		}
	}

	for i, link := range links {
		linkName := htmlquery.InnerText(link)

		vSeries := strings.HasSuffix(linkName, "(V Series)")

		if preferPremium && vSeries {
			continue
		}

		href, err := getLinkFromTag(link, linkName, searchURL, i)
		if err != nil {
			log.Warn(err)
			continue
		}

		log.Debugf("Found link %s for card %s", href, cardName)

		return href, nil
	}

	return "", fmt.Errorf("couldn't find link in %s for card %s", searchURL, cardName)
}

func innerText(node *html.Node) string {
	var value string
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode && child.Data == "noscript" {
			continue
		}
		value += htmlquery.InnerText(child)
	}
	return strings.TrimSpace(value)
}

func parseTextNode(node *html.Node) string {
	var value string
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode {
			value += htmlquery.InnerText(child)
		}
	}
	return strings.TrimSpace(value)
}

func parseTextNodeAsInt(node *html.Node) (int, error) {
	var value string
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode {
			value += htmlquery.InnerText(child)
		}
	}
	return strconv.Atoi(strings.TrimSpace(value))
}

func parseLinkNode(node *html.Node) string {
	var value string
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode {
			value += htmlquery.InnerText(child)
		} else if child.Type == html.ElementNode && child.Data == "a" {
			value += htmlquery.InnerText(child)
		}
	}
	return strings.TrimSpace(value)
}

func getOptionalStringValue(cardPage *html.Node, xpathExpr *xpath.Expr, value **string) {
	node := htmlquery.QuerySelector(cardPage, xpathExpr)
	if node != nil {
		parsedValue := parseTextNode(node)
		if len(parsedValue) > 0 {
			*value = &parsedValue
		}
	}
}

func getOptionalStringValueFromLink(cardPage *html.Node, xpathExpr *xpath.Expr, value **string) {
	node := htmlquery.QuerySelector(cardPage, xpathExpr)
	if node != nil {
		parsedValue := parseLinkNode(node)
		if len(parsedValue) > 0 {
			*value = &parsedValue
		}
	}
}

func getOptionalIntValue(cardPage *html.Node, cardPageURL string, fieldName string, xpathExpr *xpath.Expr, value **int) {
	node := htmlquery.QuerySelector(cardPage, xpathExpr)
	if node != nil {
		parsedValue, err := parseTextNodeAsInt(node)
		if err == nil {
			*value = &parsedValue
		}
	}
}

func getCardImages(cardPage *html.Node, cardPageURL string, card *Card) error {
	englishImageURL := htmlquery.QuerySelector(cardPage, englishImageURLXPath)
	if englishImageURL == nil {
		// On some pages, only the English image is available
		englishImageURL = htmlquery.QuerySelector(cardPage, defaultImageURLXPath)
		if englishImageURL == nil {
			return fmt.Errorf("no English image found in %s (XPath: %s and %s)", cardPageURL, englishImageURLXPath, defaultImageURLXPath)
		}
		card.EnglishImageURL = trimImageURL(htmlquery.InnerText(englishImageURL))
		return nil
	}
	card.EnglishImageURL = trimImageURL(htmlquery.InnerText(englishImageURL))

	japaneseImageURL := htmlquery.QuerySelector(cardPage, japaneseImageURLXPath)
	if japaneseImageURL == nil {
		return fmt.Errorf("no Japanese image found in %s (XPath: %s)", cardPageURL, japaneseImageURLXPath)
	}
	card.JapaneseImageURL = trimImageURL(htmlquery.InnerText(japaneseImageURL))

	return nil
}

func getCardNames(cardPage *html.Node, cardPageURL string, card *Card) error {
	englishName := htmlquery.QuerySelector(cardPage, englishNameXPath)
	if englishName == nil {
		return fmt.Errorf("no English name found in %s (XPath: %s)", cardPageURL, englishNameXPath)
	}
	card.EnglishName = strings.TrimSpace(htmlquery.InnerText(englishName))

	japaneseName := htmlquery.QuerySelector(cardPage, kanjiNameXPath)
	if japaneseName == nil {
		japaneseName = htmlquery.QuerySelector(cardPage, kanaNameXPath)
		if japaneseName == nil {
			return fmt.Errorf("no Japanese name found in %s (XPath: %s and %s)", cardPageURL, kanjiNameXPath, kanaNameXPath)
		}
	}
	card.JapaneseName = strings.TrimSpace(htmlquery.InnerText(japaneseName))

	return nil
}

// GetCard retrieves a card's information from https://cardfight.fandom.com/
func GetCard(cardName string, preferPremium bool) (Card, error) {
	var card Card

	cardPageURL, err := search(cardName, preferPremium)
	if err != nil {
		return card, err
	}

	cardPage, err := htmlquery.LoadURL(cardPageURL)
	if err != nil {
		return card, fmt.Errorf("couldn't query %s: %w", cardPageURL, err)
	}

	err = getCardImages(cardPage, cardPageURL, &card)
	if err != nil {
		return card, err
	}

	err = getCardNames(cardPage, cardPageURL, &card)
	if err != nil {
		return card, err
	}

	getOptionalStringValueFromLink(cardPage, cardTypeXPath, &card.Type)

	gradeSkill := htmlquery.QuerySelector(cardPage, gradeSkillXPath)
	if gradeSkill == nil {
		return card, fmt.Errorf("no grade / skill found in %s (XPath: %s)", cardPageURL, gradeSkillXPath)
	}
	split := strings.Split(innerText(gradeSkill), " / ")
	card.Grade, err = strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(split[0]), "Grade "))
	if err != nil {
		return card, fmt.Errorf("invalid grade value found in %s: %s", cardPageURL, split[0])
	}
	if len(split) > 1 {
		skill := strings.TrimSpace(split[1])
		card.Skill = &skill
	}

	getOptionalStringValue(cardPage, powerXPath, &card.Power)
	getOptionalIntValue(cardPage, cardPageURL, "critical", criticalXPath, &card.Critical)
	getOptionalIntValue(cardPage, cardPageURL, "shield", shieldXPath, &card.Shield)
	getOptionalStringValue(cardPage, nationXPath, &card.Nation)
	getOptionalStringValueFromLink(cardPage, clanXPath, &card.Clan)
	getOptionalStringValueFromLink(cardPage, raceXPath, &card.Race)
	getOptionalStringValue(cardPage, triggerEffectXPath, &card.TriggerEffect)

	formats := htmlquery.QuerySelector(cardPage, formatsXPath)
	if formats == nil {
		return card, fmt.Errorf("no format found in %s (XPath: %s)", cardPageURL, formatsXPath)
	}
	card.Formats = strings.Split(parseTextNode(formats), " / ")

	flavor := htmlquery.QuerySelector(cardPage, flavorXPath)
	if flavor != nil {
		var sb strings.Builder
		for child := flavor.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode && child.Data == "br" {
				sb.WriteString("\n")
			} else {
				sb.WriteString(htmlquery.InnerText(child))
			}
		}
		flavor := strings.TrimSpace(sb.String())
		card.Flavor = &flavor
	}

	effect := htmlquery.QuerySelector(cardPage, effectXPath)
	if effect != nil {
		var effectText string
		for child := effect.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == html.ElementNode && child.Data == "br" {
				effectText += "\n"
			} else if child.Type == html.ElementNode && child.Data == "b" {
				effectText += "[b]" + strings.TrimSpace(htmlquery.InnerText(child)) + "[/b]"
			} else if child.Type == html.ElementNode && child.Data == "i" {
				effectText += "[i]" + strings.TrimSpace(htmlquery.InnerText(child)) + "[/i]"
			} else if child.Type == html.ElementNode && child.Data == "font" {
				color := "-"
				for _, attr := range child.Attr {
					if attr.Key == "color" {
						switch attr.Val {
						case "red":
							color = "ff0000"
						}
					}
				}
				effectText += "[" + color + "]" + strings.TrimSpace(htmlquery.InnerText(child)) + "[-]"
			} else {
				effectText += htmlquery.InnerText(child)
			}
		}
		effectText = strings.TrimSpace(effectText)
		card.Effect = &effectText
	}

	return card, nil
}
