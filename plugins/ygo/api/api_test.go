package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	normalResponse = `[[{"id":"40640057","name":"Kuriboh","desc":"During your opponent's turn, at damage calculation: You can discard this card; you take no battle damage from that battle (this is a Quick Effect).","atk":"300","def":"200","type":"Effect Monster","level":"1","race":"Fiend","attribute":"DARK","scale":null,"linkval":null,"linkmarkers":null,"archetype":"Kuriboh","set_tag":"TEST1,TEST2,","setcode":"Test 1,Test 2","ban_tcg":null,"ban_ocg":null,"ban_goat":null,"image_url":"https:\/\/storage.googleapis.com\/ygoprodeck.com\/pics\/40640057.jpg","image_url_small":"https:\/\/storage.googleapis.com\/ygoprodeck.com\/pics_small\/40640057.jpg"}]]`
	linkResponse   = `[[{"id":"1861629","name":"Decode Talker","desc":"2+ Effect Monsters\r\nGains 500 ATK for each monster it points to. When your opponent activates a card or effect that targets a card(s) you control (Quick Effect): You can Tribute 1 monster this card points to; negate the activation, and if you do, destroy that card.","atk":"2300","def":null,"type":"Link Monster","level":"0","race":"Cyberse","attribute":"DARK","scale":null,"linkval":"3","linkmarkers":"Top,Bottom-Left,Bottom-Right","archetype":null,"set_tag":"YS18-EN043,YS17-EN041,SP18-EN031,OP06-EN001,","setcode":"Starter Deck: Codebreaker,Starter Deck: Link Strike,Star Pack VRAINS,OTS Tournament Pack 6,","ban_tcg":null,"ban_ocg":null,"ban_goat":null,"image_url":"https:\/\/storage.googleapis.com\/ygoprodeck.com\/pics\/1861629.jpg","image_url_small":"https:\/\/storage.googleapis.com\/ygoprodeck.com\/pics_small\/1861629.jpg"}]]`
	magicResponse  = `[[{"id":"5318639","name":"Mystical Space Typhoon","desc":"Target 1 Spell\/Trap on the field; destroy that target.","atk":null,"def":null,"type":"Spell Card","level":"0","race":"Quick-Play","attribute":"0","scale":null,"linkval":null,"linkmarkers":null,"archetype":null,"set_tag":"","setcode":"","ban_tcg":null,"ban_ocg":null,"ban_goat":"Limited","image_url":"https:\/\/storage.googleapis.com\/ygoprodeck.com\/pics\/5318639.jpg","image_url_small":"https:\/\/storage.googleapis.com\/ygoprodeck.com\/pics_small\/5318639.jpg"}]]`
	trapResponse   = `[[{"id":"4206964","name":"Trap Hole","desc":"When your opponent Normal or Flip Summons 1 monster with 1000 or more ATK: Target that monster; destroy that target.","atk":null,"def":null,"type":"Trap Card","level":"0","race":"Normal","attribute":"0","scale":null,"linkval":null,"linkmarkers":null,"archetype":"Hole","set_tag":"","setcode":"","ban_tcg":null,"ban_ocg":null,"ban_goat":null,"image_url":"https:\/\/storage.googleapis.com\/ygoprodeck.com\/pics\/4206964.jpg","image_url_small":"https:\/\/storage.googleapis.com\/ygoprodeck.com\/pics_small\/4206964.jpg"}]]`
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

	_, err := Query("1", options...)
	assert.NotNil(t, err)
}

func TestInvalidResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `[[]]`)
	})
	ts, options := setupTestServer(handler)
	defer ts.Close()

	_, err := Query("1", options...)
	assert.NotNil(t, err)
}

func TestNormal(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, normalResponse)
	})
	ts, options := setupTestServer(handler)
	defer ts.Close()

	id := "40640057"
	name := "Kuriboh"
	attack := "300"
	defense := "200"
	level := "1"
	expected := Data{
		YGOProID:      id,
		Name:          name,
		Description:   "During your opponent's turn, at damage calculation: You can discard this card; you take no battle damage from that battle (this is a Quick Effect).",
		Attack:        &attack,
		Defense:       &defense,
		Type:          TypeEffectMonster,
		Level:         &level,
		Race:          RaceFiend,
		Attribute:     AttributeDark,
		Archetype:     &name,
		SetTags:       Tags{"TEST1", "TEST2"},
		SetCodes:      Tags{"Test 1", "Test 2"},
		ImageURL:      "https://storage.googleapis.com/ygoprodeck.com/pics/" + id + ".jpg",
		ImageURLSmall: "https://storage.googleapis.com/ygoprodeck.com/pics_small/" + id + ".jpg",
	}

	resp, err := Query(id, options...)
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

	id := "1861629"
	name := "Decode Talker"
	attack := "2300"
	level := "0"
	linkValue := "3"
	expected := Data{
		YGOProID:      id,
		Name:          name,
		Description:   "2+ Effect Monsters\r\nGains 500 ATK for each monster it points to. When your opponent activates a card or effect that targets a card(s) you control (Quick Effect): You can Tribute 1 monster this card points to; negate the activation, and if you do, destroy that card.",
		Attack:        &attack,
		Type:          TypeLinkMonster,
		Level:         &level,
		Race:          RaceCyberse,
		Attribute:     AttributeDark,
		LinkValue:     &linkValue,
		LinkMarkers:   LinkMarkers{LinkMarkerTop, LinkMarkerBottomLeft, LinkMarkerBottomRight},
		SetTags:       Tags{"YS18-EN043", "YS17-EN041", "SP18-EN031", "OP06-EN001"},
		SetCodes:      Tags{"Starter Deck: Codebreaker", "Starter Deck: Link Strike", "Star Pack VRAINS", "OTS Tournament Pack 6"},
		ImageURL:      "https://storage.googleapis.com/ygoprodeck.com/pics/" + id + ".jpg",
		ImageURLSmall: "https://storage.googleapis.com/ygoprodeck.com/pics_small/" + id + ".jpg",
	}

	resp, err := Query(id, options...)
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

	id := "5318639"
	name := "Mystical Space Typhoon"
	level := "0"
	banStatus := BanStatusLimited
	expected := Data{
		YGOProID:      id,
		Name:          name,
		Description:   "Target 1 Spell/Trap on the field; destroy that target.",
		Type:          TypeSpellCard,
		Level:         &level,
		Race:          RaceQuickPlay,
		Attribute:     AttributeNone,
		SetTags:       nil,
		SetCodes:      nil,
		BanGOAT:       &banStatus,
		ImageURL:      "https://storage.googleapis.com/ygoprodeck.com/pics/" + id + ".jpg",
		ImageURLSmall: "https://storage.googleapis.com/ygoprodeck.com/pics_small/" + id + ".jpg",
	}

	resp, err := Query(id, options...)
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

	id := "4206964"
	name := "Trap Hole"
	level := "0"
	archetype := "Hole"
	expected := Data{
		YGOProID:      id,
		Name:          name,
		Description:   "When your opponent Normal or Flip Summons 1 monster with 1000 or more ATK: Target that monster; destroy that target.",
		Type:          TypeTrapCard,
		Race:          RaceNormal,
		Level:         &level,
		Attribute:     AttributeNone,
		Archetype:     &archetype,
		SetTags:       nil,
		SetCodes:      nil,
		ImageURL:      "https://storage.googleapis.com/ygoprodeck.com/pics/" + id + ".jpg",
		ImageURLSmall: "https://storage.googleapis.com/ygoprodeck.com/pics_small/" + id + ".jpg",
	}

	resp, err := Query(id, options...)
	assert.Nil(t, err)
	assert.Equal(t, expected, resp)
}
