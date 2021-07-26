package pkm

import (
	"fmt"
	"time"

	pokemontcgsdk "github.com/PokemonTCG/pokemon-tcg-sdk-go-v2/pkg"
	"github.com/PokemonTCG/pokemon-tcg-sdk-go-v2/pkg/request"
)

// See https://docs.pokemontcg.io/#documentationrate_limits
var rateLimiter = time.NewTicker(1.4 * 1000 * time.Millisecond)

func getCards(name string, setCode string) ([]pokemontcgsdk.PokemonCard, error) {
	<-rateLimiter.C
	
	name_query := fmt.Sprintf("name:%s", name)
	set_query := fmt.Sprintf("set.id:%s", setCode)
	tcg := pokemontcgsdk.NewClient("")
	
	cards, err := tcg.GetCards(
		request.Query(name_query, set_query),
		request.PageSize(5),
	)

	deref := []pokemontcgsdk.PokemonCard{}

	for _, card := range cards {
		deref = append(deref, *card)
	}
	
	return deref, err
}

func getSets() ([]pokemontcgsdk.Set, error) {
	<-rateLimiter.C
	tcg := pokemontcgsdk.NewClient("")
	
	sets, err := tcg.GetSets(
		request.PageSize(200),
	)
	deref := []pokemontcgsdk.Set{}

	for _, set := range sets {
		deref = append(deref, *set)
	}
	return deref, err
}
