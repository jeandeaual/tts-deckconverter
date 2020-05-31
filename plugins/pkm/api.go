package pkm

import (
	"time"

	pokemontcgsdk "github.com/PokemonTCG/pokemon-tcg-sdk-go/src"
)

// See https://docs.pokemontcg.io/#documentationrate_limits
var rateLimiter = time.NewTicker(1.4 * 1000 * time.Millisecond)

func getCards(name string, setCode string) ([]pokemontcgsdk.PokemonCard, error) {
	<-rateLimiter.C
	return pokemontcgsdk.GetCards(map[string]string{
		"name":    name,
		"setCode": setCode,
	})
}

func getSets() ([]pokemontcgsdk.Set, error) {
	<-rateLimiter.C
	return pokemontcgsdk.GetSets(nil)
}
