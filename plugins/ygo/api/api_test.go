package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	normalResponse = `{"data":[{"id":40640057,"name":"Kuriboh","type":"Effect Monster","desc":"During damage calculation, if your opponent's monster attacks (Quick Effect): You can discard this card; you take no battle damage from that battle.","atk":300,"def":200,"level":1,"race":"Fiend","attribute":"DARK","archetype":"Kuriboh","card_sets":[{"set_name":"Test 1","set_code":"TEST1","set_rarity":"Rare","set_price":""},{"set_name":"Test 2","set_code":"TEST2","set_rarity":"Rare","set_price":""}],"card_images":[{"id":40640057,"image_url":"https://storage.googleapis.com/ygoprodeck.com/pics/40640057.jpg","image_url_small":"https://storage.googleapis.com/ygoprodeck.com/pics_small/40640057.jpg"},{"id":40640058,"image_url":"https://storage.googleapis.com/ygoprodeck.com/pics/40640058.jpg","image_url_small":"https://storage.googleapis.com/ygoprodeck.com/pics_small/40640058.jpg"},{"id":40640059,"image_url":"https://storage.googleapis.com/ygoprodeck.com/pics/40640059.jpg","image_url_small":"https://storage.googleapis.com/ygoprodeck.com/pics_small/40640059.jpg"}],"card_prices":[]}]}`
	linkResponse   = `{"data":[{"id":1861629,"name":"Decode Talker","type":"Link Monster","desc":"2+ Effect Monsters\nGains 500 ATK for each monster it points to. When your opponent activates a card or effect that targets a card(s) you control (Quick Effect): You can Tribute 1 monster this card points to; negate the activation, and if you do, destroy that card.","atk":2300,"race":"Cyberse","attribute":"DARK","linkval":3,"linkmarkers":["Top","Bottom-Left","Bottom-Right"],"card_sets":[{"set_name":"Duel Devastator","set_code":"DUDE-EN023","set_rarity":"Ultra Rare","set_price":""},{"set_name":"Duel Power","set_code":"DUPO-EN106","set_rarity":"Ultra Rare","set_price":""},{"set_name":"OTS Tournament Pack 6","set_code":"OP06-EN001","set_rarity":"Ultimate Rare","set_price":""},{"set_name":"Star Pack VRAINS","set_code":"SP18-EN031","set_rarity":"Starfoil Rare","set_price":""},{"set_name":"Starter Deck: Codebreaker","set_code":"YS18-EN043","set_rarity":"Common","set_price":""},{"set_name":"Starter Deck: Link Strike","set_code":"YS17-EN041","set_rarity":"Ultra Rare","set_price":""}],"card_images":[{"id":1861629,"image_url":"https://storage.googleapis.com/ygoprodeck.com/pics/1861629.jpg","image_url_small":"https://storage.googleapis.com/ygoprodeck.com/pics_small/1861629.jpg"},{"id":1861630,"image_url":"https://storage.googleapis.com/ygoprodeck.com/pics/1861630.jpg","image_url_small":"https://storage.googleapis.com/ygoprodeck.com/pics_small/1861630.jpg"}],"card_prices":[]}]}`
	magicResponse  = `{"data":[{"id":5318639,"name":"Mystical Space Typhoon","type":"Spell Card","desc":"Target 1 Spell/Trap on the field; destroy that target.","race":"Quick-Play","card_sets":[],"banlist_info":{"ban_goat":"Limited"},"card_images":[{"id":5318639,"image_url":"https://storage.googleapis.com/ygoprodeck.com/pics/5318639.jpg","image_url_small":"https://storage.googleapis.com/ygoprodeck.com/pics_small/5318639.jpg"}],"card_prices":[]}]}`
	trapResponse   = `{"data":[{"id":4206964,"name":"Trap Hole","type":"Trap Card","desc":"When your opponent Normal or Flip Summons 1 monster with 1000 or more ATK: Target that monster; destroy that target.","race":"Normal","archetype":"Hole","card_sets":[],"card_images":[{"id":4206964,"image_url":"https://storage.googleapis.com/ygoprodeck.com/pics/4206964.jpg","image_url_small":"https://storage.googleapis.com/ygoprodeck.com/pics_small/4206964.jpg"}],"card_prices":[]}]}`
)

func TestType(t *testing.T) {
	assert.True(t, TypeNormalMonster.IsMonster())
	assert.True(t, TypeEffectMonster.IsMonster())
	assert.True(t, TypeLinkMonster.IsMonster())
	assert.False(t, TypeSpellCard.IsMonster())
	assert.False(t, TypeTrapCard.IsMonster())
	assert.False(t, TypeSkillCard.IsMonster())
	assert.True(t, TypeXYZMonster.IsXYZ())
	assert.True(t, TypeXYZPendulumEffectMonster.IsXYZ())
	assert.False(t, TypeFlipEffectMonster.IsXYZ())
}

func setupTestServer(handler func(http.ResponseWriter, *http.Request)) (*httptest.Server, []ClientOption) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	ts := httptest.NewServer(mux)

	return ts, []ClientOption{WithBaseURL(ts.URL)}
}

func TestNotFound(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	ts, options := setupTestServer(handler)
	defer ts.Close()

	_, err := QueryID(1, FormatStandard, options...)
	assert.NotNil(t, err)
}

func TestInvalidResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `[[]]`)
	})
	ts, options := setupTestServer(handler)
	defer ts.Close()

	_, err := QueryID(1, FormatStandard, options...)
	assert.NotNil(t, err)
}

func TestNormal(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, normalResponse)
	})
	ts, options := setupTestServer(handler)
	defer ts.Close()

	var id int64 = 40640057
	name := "Kuriboh"
	attack := 300
	defense := 200
	level := 1
	attribute := AttributeDark
	expected := Data{
		YGOProID:    id,
		Name:        name,
		Description: "During damage calculation, if your opponent's monster attacks (Quick Effect): You can discard this card; you take no battle damage from that battle.",
		Attack:      &attack,
		Defense:     &defense,
		Type:        TypeEffectMonster,
		Level:       &level,
		Race:        RaceFiend,
		Attribute:   &attribute,
		Archetype:   &name,
		Sets: []CardSet{
			{
				Code:   "TEST1",
				Name:   "Test 1",
				Rarity: "Rare",
			},
			{
				Code:   "TEST2",
				Name:   "Test 2",
				Rarity: "Rare",
			},
		},
		Images: []CardImage{
			{
				ID:       id,
				URL:      fmt.Sprintf("https://storage.googleapis.com/ygoprodeck.com/pics/%d.jpg", id),
				URLSmall: fmt.Sprintf("https://storage.googleapis.com/ygoprodeck.com/pics_small/%d.jpg", id),
			},
			{
				ID:       id + 1,
				URL:      fmt.Sprintf("https://storage.googleapis.com/ygoprodeck.com/pics/%d.jpg", id+1),
				URLSmall: fmt.Sprintf("https://storage.googleapis.com/ygoprodeck.com/pics_small/%d.jpg", id+1),
			},
			{
				ID:       id + 2,
				URL:      fmt.Sprintf("https://storage.googleapis.com/ygoprodeck.com/pics/%d.jpg", id+2),
				URLSmall: fmt.Sprintf("https://storage.googleapis.com/ygoprodeck.com/pics_small/%d.jpg", id+2),
			},
		},
		Prices: []CardPrice{},
	}

	resp, err := QueryID(id, FormatStandard, options...)
	assert.Nil(t, err)
	assert.Equal(t, expected, resp)
}

func TestLink(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, linkResponse)
	})
	ts, options := setupTestServer(handler)
	defer ts.Close()

	var id int64 = 1861629
	name := "Decode Talker"
	attack := 2300
	attribute := AttributeDark
	linkValue := 3
	expected := Data{
		YGOProID:    id,
		Name:        name,
		Description: "2+ Effect Monsters\nGains 500 ATK for each monster it points to. When your opponent activates a card or effect that targets a card(s) you control (Quick Effect): You can Tribute 1 monster this card points to; negate the activation, and if you do, destroy that card.",
		Attack:      &attack,
		Type:        TypeLinkMonster,
		Race:        RaceCyberse,
		Attribute:   &attribute,
		LinkValue:   &linkValue,
		LinkMarkers: []LinkMarker{LinkMarkerTop, LinkMarkerBottomLeft, LinkMarkerBottomRight},
		Sets: []CardSet{
			{
				Code:   "DUDE-EN023",
				Name:   "Duel Devastator",
				Rarity: "Ultra Rare",
			},
			{
				Code:   "DUPO-EN106",
				Name:   "Duel Power",
				Rarity: "Ultra Rare",
			},
			{
				Code:   "OP06-EN001",
				Name:   "OTS Tournament Pack 6",
				Rarity: "Ultimate Rare",
			},
			{
				Code:   "SP18-EN031",
				Name:   "Star Pack VRAINS",
				Rarity: "Starfoil Rare",
			},
			{
				Code:   "YS18-EN043",
				Name:   "Starter Deck: Codebreaker",
				Rarity: "Common",
			},
			{
				Code:   "YS17-EN041",
				Name:   "Starter Deck: Link Strike",
				Rarity: "Ultra Rare",
			},
		},
		Images: []CardImage{
			{
				ID:       id,
				URL:      fmt.Sprintf("https://storage.googleapis.com/ygoprodeck.com/pics/%d.jpg", id),
				URLSmall: fmt.Sprintf("https://storage.googleapis.com/ygoprodeck.com/pics_small/%d.jpg", id),
			},
			{
				ID:       id + 1,
				URL:      fmt.Sprintf("https://storage.googleapis.com/ygoprodeck.com/pics/%d.jpg", id+1),
				URLSmall: fmt.Sprintf("https://storage.googleapis.com/ygoprodeck.com/pics_small/%d.jpg", id+1),
			},
		},
		Prices: []CardPrice{},
	}

	resp, err := QueryID(id, FormatStandard, options...)
	assert.Nil(t, err)
	assert.Equal(t, expected, resp)
}

func TestMagic(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, magicResponse)
	})
	ts, options := setupTestServer(handler)
	defer ts.Close()

	var id int64 = 5318639
	name := "Mystical Space Typhoon"
	banStatus := BanStatusLimited
	expected := Data{
		YGOProID:    id,
		Name:        name,
		Description: "Target 1 Spell/Trap on the field; destroy that target.",
		Type:        TypeSpellCard,
		Race:        RaceQuickPlay,
		BanListInfo: &BanListInfo{
			BanGOAT: &banStatus,
		},
		Sets: []CardSet{},
		Images: []CardImage{
			{
				ID:       id,
				URL:      fmt.Sprintf("https://storage.googleapis.com/ygoprodeck.com/pics/%d.jpg", id),
				URLSmall: fmt.Sprintf("https://storage.googleapis.com/ygoprodeck.com/pics_small/%d.jpg", id),
			},
		},
		Prices: []CardPrice{},
	}

	resp, err := QueryID(id, FormatStandard, options...)
	assert.Nil(t, err)
	assert.Equal(t, expected, resp)
}

func TestTrap(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, trapResponse)
	})
	ts, options := setupTestServer(handler)
	defer ts.Close()

	var id int64 = 4206964
	name := "Trap Hole"
	archetype := "Hole"
	expected := Data{
		YGOProID:    id,
		Name:        name,
		Description: "When your opponent Normal or Flip Summons 1 monster with 1000 or more ATK: Target that monster; destroy that target.",
		Type:        TypeTrapCard,
		Race:        RaceNormal,
		Archetype:   &archetype,
		Sets:        []CardSet{},
		Images: []CardImage{
			{
				ID:       id,
				URL:      fmt.Sprintf("https://storage.googleapis.com/ygoprodeck.com/pics/%d.jpg", id),
				URLSmall: fmt.Sprintf("https://storage.googleapis.com/ygoprodeck.com/pics_small/%d.jpg", id),
			},
		},
		Prices: []CardPrice{},
	}

	resp, err := QueryID(id, FormatStandard, options...)
	assert.Nil(t, err)
	assert.Equal(t, expected, resp)
}
