package vanguard

import (
	"time"

	"github.com/jeandeaual/tts-deckconverter/plugins/vanguard/cardfightwiki"
)

var rateLimiter = time.NewTicker(100 * time.Millisecond)

func getCard(name string, preferPremium bool) (cardfightwiki.Card, error) {
	<-rateLimiter.C
	return cardfightwiki.GetCard(name, preferPremium)
}
