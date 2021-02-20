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

func buildCardName(card scryfall.Card) string {
	var sb strings.Builder

	sb.WriteString(card.Name)
	sb.WriteString("\n")

	if card.CMC > 0 {
		sb.WriteString(strconv.FormatFloat(card.CMC, 'f', -1, 64))
		sb.WriteString("CMC")
		sb.WriteString("\n")
	}

	sb.WriteString("[b]")
	sb.WriteString(card.TypeLine)
	sb.WriteString("[/b]")

	return sb.String()
}

func buildCardDescription(card scryfall.Card, rulings []scryfall.Ruling, detailedDescription bool) string {
	var sb strings.Builder

	if !detailedDescription {
		if card.Power != nil && card.Toughness != nil {
			sb.WriteString("[b]")
			sb.WriteString(*card.Power)
			sb.WriteString("/")
			sb.WriteString(*card.Toughness)
			sb.WriteString("[/b]")
		} else if card.Loyalty != nil {
			sb.WriteString("[b]")
			sb.WriteString(*card.Loyalty)
			sb.WriteString("[/b]")
		} else {
			return ""
		}
	}

	if len(card.ManaCost) > 0 {
		sb.WriteString(card.ManaCost)
	}

	if len(card.OracleText) > 0 {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(card.OracleText)
	}

	if card.FlavorText != nil {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString("[i]")
		sb.WriteString(*card.FlavorText)
		sb.WriteString("[/i]")
	}

	if card.Power != nil && card.Toughness != nil {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString("[b]")
		sb.WriteString(*card.Power)
		sb.WriteString("/")
		sb.WriteString(*card.Toughness)
		sb.WriteString("[/b]")
	} else if card.Loyalty != nil {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString("Loyalty: [b]")
		sb.WriteString(*card.Loyalty)
		sb.WriteString("[/b]")
	} else if card.HandModifier != nil && card.LifeModifier != nil {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString("Hand modifier: [b]")
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

func buildCardFaceName(name string, cmc float64, typeLine string) string {
	var sb strings.Builder

	sb.WriteString(name)
	sb.WriteString("\n")

	if cmc > 0 {
		sb.WriteString(strconv.FormatFloat(cmc, 'f', -1, 64))
		sb.WriteString("CMC")
		sb.WriteString("\n")
	}

	sb.WriteString("[b]")
	sb.WriteString(typeLine)
	sb.WriteString("[/b]")

	return sb.String()
}

func buildCardFacesName(card scryfall.Card) string {
	if len(card.CardFaces) == 0 {
		return ""
	}

	var (
		name     strings.Builder
		typeLine strings.Builder
	)

	for i, face := range card.CardFaces {
		name.WriteString(face.Name)
		typeLine.WriteString(face.TypeLine)

		if i < len(card.CardFaces)-1 {
			name.WriteString(" // ")
			typeLine.WriteString(" // ")
		}
	}

	return buildCardFaceName(name.String(), card.CMC, typeLine.String())
}

func buildCardFaceDescription(face scryfall.CardFace, rulings []scryfall.Ruling, detailedDescription bool) string {
	var sb strings.Builder

	if !detailedDescription {
		if face.Power != nil && face.Toughness != nil {
			sb.WriteString("[b]")
			sb.WriteString(*face.Power)
			sb.WriteString("/")
			sb.WriteString(*face.Toughness)
			sb.WriteString("[/b]")
		} else if face.Loyalty != nil {
			sb.WriteString("[b]")
			sb.WriteString(*face.Loyalty)
			sb.WriteString("[/b]")
		} else {
			return ""
		}
	}

	if len(face.ManaCost) > 0 {
		sb.WriteString(face.ManaCost)
	}

	if face.OracleText != nil {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(*face.OracleText)
	}

	if face.FlavorText != nil {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString("[i]")
		sb.WriteString(*face.FlavorText)
		sb.WriteString("[/i]")
	}

	if face.Power != nil && face.Toughness != nil {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString("[b]")
		sb.WriteString(*face.Power)
		sb.WriteString("/")
		sb.WriteString(*face.Toughness)
		sb.WriteString("[/b]")
	} else if face.Loyalty != nil {
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString("Loyalty: [b]")
		sb.WriteString(*face.Loyalty)
		sb.WriteString("[/b]")
	}

	if rulings != nil {
		appendRulings(&sb, rulings)
	}

	return sb.String()
}

func buildCardFacesDescription(faces []scryfall.CardFace, rulings []scryfall.Ruling, detailedDescription bool) string {
	if len(faces) == 0 {
		return ""
	}

	var sb strings.Builder

	for i, face := range faces {
		sb.WriteString("[u]")
		sb.WriteString(face.Name)
		sb.WriteString("[/u]\n\n")
		sb.WriteString(buildCardFaceDescription(face, nil, detailedDescription))

		if i < len(faces)-1 {
			sb.WriteString("\n\n--------------------------\n\n")
		}
	}

	if rulings != nil {
		appendRulings(&sb, rulings)
	}

	return sb.String()
}
