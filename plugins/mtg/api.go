package mtg

import (
	"context"
	"time"

	scryfall "github.com/BlueMonday/go-scryfall"
)

// See https://scryfall.com/docs/api#rate-limits-and-good-citizenship
var rateLimiter = time.NewTicker(100 * time.Millisecond)

func getCard(ctx context.Context, client *scryfall.Client, id string) (scryfall.Card, error) {
	<-rateLimiter.C
	return client.GetCard(ctx, id)
}

func getCardByName(ctx context.Context, client *scryfall.Client, name string, opts scryfall.GetCardByNameOptions) (scryfall.Card, error) {
	<-rateLimiter.C
	// Fuzzy search is required to match card names in languages other
	// than English ("printed_name")
	return client.GetCardByName(ctx, name, false, opts)
}

func listSets(ctx context.Context, client *scryfall.Client) ([]scryfall.Set, error) {
	<-rateLimiter.C
	return client.ListSets(ctx)
}

func getRulings(ctx context.Context, client *scryfall.Client, cardID string) ([]scryfall.Ruling, error) {
	<-rateLimiter.C
	return client.GetRulings(ctx, cardID)
}
