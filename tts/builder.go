package tts

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strconv"

	"go.uber.org/zap"

	"deckconverter/plugins"
)

const (
	// Default cards are 56×79mm for some reason, so we need to scale them
	// to get the correct size (63×88mm)
	standardScaleX = 63.5 / 56
	standardScaleZ = 88.9 / 79
	// Oversized cards (MTG Planes, Schemes, Vanguards) are approximatively
	// 88×124mm
	standardOversizedScale = 88.9 / 63.5
	// With scale 1.0, small cards are 58×79mm, so we need to scale them
	// to get the correct size (59×86mm)
	smallScaleX = 59.0 / 58
	smallScaleZ = 86.0 / 79
)

func createDeck(deck *plugins.Deck, log *zap.SugaredLogger) (SavedObject, string) {
	object := createDefaultDeck()
	count := 1
	thumbnailSource := ""
	deckObject := &object.ObjectStates[0]
	oversizedDeck := true

	for _, card := range deck.Cards {
		var (
			customDeck CustomDeck
			template   *plugins.Template
		)

		if deck.TemplateInfo == nil {
			customDeck = CustomDeck{
				FaceURL:      card.ImageURL,
				BackURL:      deck.BackURL,
				NumWidth:     1,
				NumHeight:    1,
				BackIsHidden: true,
				UniqueBack:   false,
			}
		} else {
			var (
				templateID int
				err        error
			)
			cardID, found := deck.TemplateInfo.ImageURLCardIDMap[card.ImageURL]
			if !found {
				log.Errorw(
					"Image ID for not found for URL",
					"url", card.ImageURL,
					"urlIDMap", deck.TemplateInfo.ImageURLCardIDMap,
				)
			}
			template, templateID, err = deck.TemplateInfo.GetAssociatedTemplate(cardID)
			if err != nil {
				log.Errorw(
					"Couldn't find template for card",
					"cardID", cardID,
					"error", err,
					"templates", deck.TemplateInfo.Templates,
					"urlIDMap", deck.TemplateInfo.ImageURLCardIDMap,
				)
			}
			customDeck = CustomDeck{
				FaceURL:      template.URL,
				BackURL:      deck.BackURL,
				NumWidth:     template.NumCols,
				NumHeight:    template.NumRows,
				BackIsHidden: true,
				UniqueBack:   false,
			}
			for i := 0; i < card.Count; i++ {
				deckObject.DeckIDs = append(deckObject.DeckIDs, cardID)
			}
			deckObject.CustomDeck[strconv.Itoa(templateID)] = customDeck
		}

		for i := 0; i < card.Count; i++ {
			if deck.TemplateInfo == nil {
				deckObject.DeckIDs = append(deckObject.DeckIDs, 100*count)
				deckObject.CustomDeck[strconv.Itoa(count)] = customDeck
			}

			if len(thumbnailSource) == 0 {
				thumbnailSource = card.ImageURL
			}

			deckObject.ContainedObjects = append(
				deckObject.ContainedObjects,
				createCard(card, count, customDeck, deck.TemplateInfo, deck.CardSize, log),
			)

			if deck.TemplateInfo == nil {
				count++
			}
		}

		if oversizedDeck && !card.Oversized {
			oversizedDeck = false
		}
	}

	switch deck.CardSize {
	case plugins.CardSizeStandard:
		deckObject.Transform.ScaleX = standardScaleX
		deckObject.Transform.ScaleZ = standardScaleZ

		if oversizedDeck {
			deckObject.Transform.ScaleX *= standardOversizedScale
			deckObject.Transform.ScaleY *= standardOversizedScale
			deckObject.Transform.ScaleZ *= standardOversizedScale
		}
	case plugins.CardSizeSmall:
		deckObject.Transform.ScaleX = smallScaleX
		deckObject.Transform.ScaleZ = smallScaleZ
	}

	return object, thumbnailSource
}

func createCard(card plugins.CardInfo, count int, customDeck CustomDeck,
	templateInfo *plugins.TemplateInfo, cardSize plugins.CardSize,
	log *zap.SugaredLogger) Object {

	var states map[string]Object

	if card.AlternativeState != nil {
		var alternateCustomDeck CustomDeck
		if templateInfo == nil {
			alternateCustomDeck = CustomDeck{
				FaceURL:      card.AlternativeState.ImageURL,
				BackURL:      customDeck.BackURL,
				NumWidth:     1,
				NumHeight:    1,
				BackIsHidden: true,
				UniqueBack:   false,
			}
		} else {
			cardID, found := templateInfo.ImageURLCardIDMap[card.AlternativeState.ImageURL]
			if !found {
				log.Errorw(
					"Image ID for not found for URL",
					"url", card.AlternativeState.ImageURL,
					"urlIDMap", templateInfo.ImageURLCardIDMap,
				)
			}
			template, _, err := templateInfo.GetAssociatedTemplate(cardID)
			if err != nil {
				log.Errorw(
					"Template for card ID",
					"cardID", cardID,
					"urlIDMap", templateInfo.ImageURLCardIDMap,
				)
			}
			alternateCustomDeck = CustomDeck{
				FaceURL:      template.URL,
				BackURL:      customDeck.BackURL,
				NumWidth:     template.NumCols,
				NumHeight:    template.NumRows,
				BackIsHidden: true,
				UniqueBack:   false,
			}
		}
		alternateState := createCard(*card.AlternativeState, 1, alternateCustomDeck, templateInfo, cardSize, log)
		states = map[string]Object{
			"2": alternateState,
		}
	}

	var (
		cardID       int
		customDeckID string
	)
	if templateInfo == nil {
		cardID = 100 * count
		customDeckID = strconv.Itoa(count)
	} else {
		var found bool

		cardID, found = templateInfo.ImageURLCardIDMap[card.ImageURL]
		if !found {
			log.Errorw(
				"Image ID for not found for URL",
				"url", card.ImageURL,
				"urlIDMap", templateInfo.ImageURLCardIDMap,
			)
		}
		_, templateID, err := templateInfo.GetAssociatedTemplate(cardID)
		if err != nil {
			log.Errorw(
				"Template for card ID",
				"cardID", cardID,
				"urlIDMap", templateInfo.ImageURLCardIDMap,
			)
		}
		customDeckID = strconv.Itoa(templateID)
	}

	scaleX := Decimal(1)
	scaleY := Decimal(1)
	scaleZ := Decimal(1)

	switch cardSize {
	case plugins.CardSizeStandard:
		scaleX = standardScaleX
		scaleZ = standardScaleZ

		if card.Oversized {
			scaleX *= standardOversizedScale
			scaleY *= standardOversizedScale
			scaleZ *= standardOversizedScale
		}
	case plugins.CardSizeSmall:
		scaleX = smallScaleX
		scaleZ = smallScaleZ
	}

	return Object{
		ObjectType:  CardObject,
		Nickname:    card.Name,
		Description: card.Description,
		Transform: Transform{
			PosX:   0,
			PosY:   0,
			PosZ:   0,
			RotX:   0,
			RotY:   180,
			RotZ:   0,
			ScaleX: scaleX,
			ScaleY: scaleY,
			ScaleZ: scaleZ,
		},
		ColorDiffuse:     DefaultColorDiffuse,
		Locked:           false,
		Grid:             true,
		Snap:             true,
		IgnoreFoW:        false,
		Autoraise:        true,
		Sticky:           true,
		Tooltip:          true,
		GridProjection:   false,
		HideWhenFaceDown: true,
		Hands:            true,
		CardID:           cardID,
		SidewaysCard:     false,
		CustomDeck: map[string]CustomDeck{
			customDeckID: customDeck,
		},
		States: states,
	}
}

func create(deck *plugins.Deck, outputFolder string, log *zap.SugaredLogger) {
	var (
		object          SavedObject
		thumbnailSource string
	)

	if len(deck.Cards) == 1 {
		// Don't create a deck, only generate a single card
		card := deck.Cards[0]
		var customDeck CustomDeck
		if deck.TemplateInfo == nil {
			customDeck = CustomDeck{
				FaceURL:      card.ImageURL,
				BackURL:      deck.BackURL,
				NumWidth:     1,
				NumHeight:    1,
				BackIsHidden: true,
				UniqueBack:   false,
			}
		} else {
			cardID, found := deck.TemplateInfo.ImageURLCardIDMap[card.ImageURL]
			if !found {
				log.Errorw(
					"Image ID for not found for URL",
					"url", card.ImageURL,
					"urlIDMap", deck.TemplateInfo.ImageURLCardIDMap,
				)
			}
			template, _, err := deck.TemplateInfo.GetAssociatedTemplate(cardID)
			if err != nil {
				log.Errorw(
					"Template for card ID",
					"cardID", cardID,
					"urlIDMap", deck.TemplateInfo.ImageURLCardIDMap,
				)
			}
			customDeck = CustomDeck{
				FaceURL:      template.URL,
				BackURL:      deck.BackURL,
				NumWidth:     template.NumCols,
				NumHeight:    template.NumRows,
				BackIsHidden: true,
				UniqueBack:   false,
			}
		}
		object = createSavedObject([]Object{
			createCard(card, 1, customDeck, deck.TemplateInfo, deck.CardSize, log),
		})
		thumbnailSource = card.ImageURL
	} else {
		object, thumbnailSource = createDeck(deck, log)
	}

	data, err := json.MarshalIndent(object, "", "  ")
	if err != nil {
		log.Error(err)
	}

	filename := filepath.Join(outputFolder, deck.Name+".json")
	log.Infof("Generating %s", filename)

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		log.Error(err)
	}

	if len(thumbnailSource) > 0 {
		downloadAndCreateThumbnail(thumbnailSource, filepath.Join(outputFolder, deck.Name+".png"), log)
	}
}

// Generate deck files inside outputFolder.
func Generate(decks []*plugins.Deck, backURL, outputFolder string, log *zap.SugaredLogger) {
	log.Infof("Generated %d decks", len(decks))

	for _, deck := range decks {
		if len(backURL) > 0 {
			deck.BackURL = backURL
		}
		create(deck, outputFolder, log)
	}
}
