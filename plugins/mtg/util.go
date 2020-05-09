package mtg

import (
	"strconv"
	"strings"

	scryfall "github.com/BlueMonday/go-scryfall"
)

const dateFormat = "2006-01-02"

func appendRulings(sb *strings.Builder, rulings []scryfall.Ruling) {
	if sb == nil || rulings == nil || len(rulings) == 0 {
		return
	}

	sb.WriteString("\n\n==========================")
	for _, ruling := range rulings {
		sb.WriteString("\n\n")
		sb.WriteString(ruling.Comment)
		sb.WriteString(" [i](")
		switch ruling.Source {
		case scryfall.SourceWOTC:
			sb.WriteString("WotC")
		default:
			sb.WriteString(strings.Title(string(ruling.Source)))
		}
		sb.WriteString(" - ")
		sb.WriteString(ruling.PublishedAt.Format(dateFormat))
		sb.WriteString(")[/i]")
	}
}

func buildCardDescription(card scryfall.Card, rulings []scryfall.Ruling) string {
	var sb strings.Builder

	if len(card.ManaCost) > 0 {
		sb.WriteString(card.ManaCost)
		sb.WriteString("\n")
	}

	if card.CMC > 0 {
		sb.WriteString("CMC ")
		sb.WriteString(strconv.FormatFloat(card.CMC, 'f', -1, 64))
		sb.WriteString("\n\n")
	}

	sb.WriteString("[b]")
	sb.WriteString(card.TypeLine)
	sb.WriteString("[/b]\n\n")

	sb.WriteString(card.OracleText)

	if card.FlavorText != nil {
		sb.WriteString("\n\n[i]")
		sb.WriteString(*card.FlavorText)
		sb.WriteString("[/i]")
	}

	if card.Power != nil && card.Toughness != nil {
		sb.WriteString("\n\n[b]")
		sb.WriteString(*card.Power)
		sb.WriteString("/")
		sb.WriteString(*card.Toughness)
		sb.WriteString("[/b]")
	} else if card.Loyalty != nil {
		sb.WriteString("\n\nLoyalty: [b]")
		sb.WriteString(*card.Loyalty)
		sb.WriteString("[/b]")
	} else if card.HandModifier != nil && card.LifeModifier != nil {
		sb.WriteString("\n\nHand modifier: [b]")
		sb.WriteString(*card.HandModifier)
		sb.WriteString("[/b]")
		sb.WriteString("\nLife modifier: [b]")
		sb.WriteString(*card.LifeModifier)
		sb.WriteString("[/b]")
	}

	if rulings != nil {
		appendRulings(&sb, rulings)
	}

	return sb.String()
}

func buildCardFaceDescription(face scryfall.CardFace, rulings []scryfall.Ruling) string {
	var sb strings.Builder

	if len(face.ManaCost) > 0 {
		sb.WriteString(face.ManaCost)
		sb.WriteString("\n\n")
	}

	sb.WriteString("[b]")
	sb.WriteString(face.TypeLine)
	sb.WriteString("[/b]\n\n")

	if face.OracleText != nil {
		sb.WriteString(*face.OracleText)
	}

	if face.FlavorText != nil {
		sb.WriteString("\n\n[i]")
		sb.WriteString(*face.FlavorText)
		sb.WriteString("[/i]")
	}

	if face.Power != nil && face.Toughness != nil {
		sb.WriteString("\n\n[b]")
		sb.WriteString(*face.Power)
		sb.WriteString("/")
		sb.WriteString(*face.Toughness)
		sb.WriteString("[/b]")
	} else if face.Loyalty != nil {
		sb.WriteString("\n\nLoyalty: [b]")
		sb.WriteString(*face.Loyalty)
		sb.WriteString("[/b]")
	}

	if rulings != nil {
		appendRulings(&sb, rulings)
	}

	return sb.String()
}

func buildCardFacesDescription(faces []scryfall.CardFace, rulings []scryfall.Ruling) string {
	if len(faces) == 0 {
		return ""
	}

	var sb strings.Builder

	for i, face := range faces {
		sb.WriteString("[u]")
		sb.WriteString(face.Name)
		sb.WriteString("[/u]\n\n")
		sb.WriteString(buildCardFaceDescription(face, nil))

		if i < len(faces)-1 {
			sb.WriteString("\n\n--------------------------\n\n")
		}
	}

	if rulings != nil {
		appendRulings(&sb, rulings)
	}

	return sb.String()
}
