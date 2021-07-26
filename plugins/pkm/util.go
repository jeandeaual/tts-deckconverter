package pkm

import (
	"strconv"
	"strings"

	pokemontcgsdk "github.com/PokemonTCG/pokemon-tcg-sdk-go-v2/pkg"
)

func formatElement(element string) string {
	switch element {
	case "Grass":
		return "[33bb33]" + element + "[ffffff]"
	case "Fire":
		return "[ff4040]" + element + "[ffffff]"
	case "Water":
		return "[00aaff]" + element + "[ffffff]"
	case "Lightning":
		return "[ffee00]" + element + "[ffffff]"
	case "Psychic":
		return "[cc00dd]" + element + "[ffffff]"
	case "Fighting":
		return "[cc7722]" + element + "[ffffff]"
	case "Darkness":
		return "[333333]" + element + "[ffffff]"
	case "Metal":
		return "[c0c0c0]" + element + "[ffffff]"
	case "Fairy":
		return "[e03a83]" + element + "[ffffff]"
	case "Dragon":
		return "[ae9962]" + element + "[ffffff]"
	default:
		return element
	}
}

func buildCost(cost []string) string {
	var (
		sb       strings.Builder
		previous string
		count    int
	)

	for _, element := range cost {
		if len(previous) == 0 {
			previous = element
			continue
		}

		count++

		if element == previous {
			continue
		}

		sb.WriteString(formatElement(previous))
		sb.WriteString("×")
		sb.WriteString(strconv.Itoa(count))
		sb.WriteString(" ")

		count = 0
		previous = element
	}

	if len(previous) > 0 {
		sb.WriteString(formatElement(previous))
		sb.WriteString("×")
		sb.WriteString(strconv.Itoa(count + 1))
	}

	return sb.String()
}

func buildCardDescription(card pokemontcgsdk.PokemonCard) string {
	var sb strings.Builder

	sb.WriteString(card.Supertype)
	sb.WriteString("\n")

	for _, subtype := range card.Subtypes {
		sb.WriteString(subtype)
		sb.WriteString(",")
	}
	if len(card.EvolvesFrom) > 0 {
		sb.WriteString(" - Evolves from ")
		sb.WriteString(card.EvolvesFrom)
	}
	sb.WriteString("\n\n")

	if len(card.Hp) > 0 && card.Hp != "None" {
		sb.WriteString(card.Hp)
		sb.WriteString(" HP")
		sb.WriteString("\n\n")
	}

	if len(card.Types) > 0 {
		for i, cardType := range card.Types {
			sb.WriteString(formatElement(cardType))
			if i < len(card.Types)-1 {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n\n")
	}

	for i, ability := range card.Abilities {
		if len(ability.Type) > 0 && len(ability.Name) > 0 && len(ability.Text) > 0 {
			sb.WriteString(ability.Type)
			sb.WriteString(": ")
			sb.WriteString("[b]")
			sb.WriteString(ability.Name)
			sb.WriteString("[/b]\n")
			sb.WriteString(ability.Text)
			if i < len(card.Attacks)-1 {
				sb.WriteString("\n\n")
			} else {
				sb.WriteString("\n")
			}
		}
	}

	for i, attack := range card.Attacks {
		sb.WriteString(buildCost(attack.Cost))
		sb.WriteString(" - ")
		sb.WriteString("[b]")
		sb.WriteString(attack.Name)
		sb.WriteString("[/b]")
		if len(attack.Damage) > 0 {
			sb.WriteString(" - ")
			sb.WriteString(attack.Damage)
		}
		if len(attack.Text) > 0 {
			sb.WriteString("\n")
			sb.WriteString(attack.Text)
		}
		if i < len(card.Attacks)-1 {
			sb.WriteString("\n\n")
		}
	}

	/* Not implementend in v2 as of 6499df97
	for i, text := range card.rules {
		sb.WriteString(text)
		if i < len(card.rules)-1 {
			sb.WriteString("\n\n")
		}
	}
	*/

	if len(card.Weaknesses) > 0 {
		sb.WriteString("\n\nResistances:\n")
		for i, weakness := range card.Weaknesses {
			sb.WriteString(formatElement(weakness.Type))
			sb.WriteString(" ")
			sb.WriteString(weakness.Value)
			if i < len(card.Weaknesses)-1 {
				sb.WriteString("\n")
			}
		}
	}

	if len(card.Resistances) > 0 {
		sb.WriteString("\n\nWeaknesses:\n")
		for i, resistance := range card.Resistances {
			sb.WriteString(formatElement(resistance.Type))
			sb.WriteString(" ")
			sb.WriteString(resistance.Value)
			if i < len(card.Resistances)-1 {
				sb.WriteString("\n")
			}
		}
	}

	if len(card.RetreatCost) > 0 {
		sb.WriteString("\n\nRetreat cost: ")
		sb.WriteString(buildCost(card.RetreatCost))
	}

	return sb.String()
}
