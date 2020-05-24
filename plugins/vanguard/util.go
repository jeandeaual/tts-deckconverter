package vanguard

import (
	"strconv"
	"strings"

	"github.com/jeandeaual/tts-deckconverter/plugins/vanguard/cardfightwiki"
)

func buildCardDescription(card cardfightwiki.Card) string {
	var sb strings.Builder

	sb.WriteString("Grade [b]")
	sb.WriteString(strconv.Itoa(card.Grade))
	sb.WriteString("[/b]\n")
	if card.Skill != nil {
		sb.WriteString("[b]")
		sb.WriteString(*card.Skill)
		sb.WriteString("[/b]")
		sb.WriteString("\n")
	}
	if card.Power != nil {
		sb.WriteString("\nPower: [b]")
		sb.WriteString(*card.Power)
		sb.WriteString("[/b]")
	}
	if card.Critical != nil {
		sb.WriteString("\nCritical: [b]")
		sb.WriteString(strconv.Itoa(*card.Critical))
		sb.WriteString("[/b]")
	}
	if card.Shield != nil {
		sb.WriteString("\nShield: [b]")
		sb.WriteString(strconv.Itoa(*card.Shield))
		sb.WriteString("[/b]")
	}
	if card.TriggerEffect != nil {
		sb.WriteString("\nTrigger Effect: [b]")
		sb.WriteString(*card.TriggerEffect)
		sb.WriteString("[/b]")
	}
	if card.Race != nil {
		sb.WriteString("\n\nRace: ")
		sb.WriteString(*card.Race)
	}
	if card.Nation != nil {
		sb.WriteString("\nNation: ")
		sb.WriteString(*card.Nation)
	}
	if card.Clan != nil {
		sb.WriteString("\nClan: ")
		sb.WriteString(*card.Clan)
	}
	if card.Effect != nil {
		sb.WriteString("\n\n")
		sb.WriteString(*card.Effect)
	}
	if card.Flavor != nil {
		sb.WriteString("\n\n[i]")
		sb.WriteString(*card.Flavor)
		sb.WriteString("[/i]")
	}

	return sb.String()
}
