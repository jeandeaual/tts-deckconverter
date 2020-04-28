package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://db.ygoprodeck.com/api/v6/cardinfo.php"
	defaultTimeout = 30 * time.Second
)

// Attribute is a card attribute.
type Attribute string

const (
	// AttributeDark represents the Dark monster card attribute.
	AttributeDark Attribute = "DARK"
	// AttributeDivine represents the Divine monster card attribute.
	AttributeDivine Attribute = "DIVINE"
	// AttributeEarth represents the Earth monster card attribute.
	AttributeEarth Attribute = "EARTH"
	// AttributeFire represents the Fire monster card attribute.
	AttributeFire Attribute = "FIRE"
	// AttributeLight represents the Light monster card attribute.
	AttributeLight Attribute = "LIGHT"
	// AttributeWater represents the Water monster card attribute.
	AttributeWater Attribute = "WATER"
	// AttributeWind represents the Wind monster card attribute.
	AttributeWind Attribute = "WIND"
	// AttributeLaugh represents the Laugh (unofficial) monster card attribute.
	AttributeLaugh Attribute = "LAUGH"
)

// Race of a card (includes spell and trap types).
type Race string

const (
	// RaceAqua is the "Aqua" monster race.
	RaceAqua Race = "Aqua"
	// RaceBeast is the "Beast" monster race.
	RaceBeast Race = "Beast"
	// RaceBeastWarrior is the "Beast-Warrior" monster race.
	RaceBeastWarrior Race = "Beast-Warrior"
	// RaceCreatorGod is the "Creator-God" monster race.
	RaceCreatorGod Race = "Creator-God"
	// RaceCyberse is the "Cyberse" monster race.
	RaceCyberse Race = "Cyberse"
	// RaceDinosaur is the "Dinosaur" monster race.
	RaceDinosaur Race = "Dinosaur"
	// RaceDivineBeast is the "Divine-Beast" monster race.
	RaceDivineBeast Race = "Divine-Beast"
	// RaceDragon is the "Dragon" monster race.
	RaceDragon Race = "Dragon"
	// RaceFairy is the "Fairy" monster race.
	RaceFairy Race = "Fairy"
	// RaceFiend is the "Fiend" monster race.
	RaceFiend Race = "Fiend"
	// RaceFish is the "Fish" monster race.
	RaceFish Race = "Fish"
	// RaceInsect is the "Insect" monster race.
	RaceInsect Race = "Insect"
	// RaceMachine is the "Machine" monster race.
	RaceMachine Race = "Machine"
	// RacePlant is the "Plant" monster race.
	RacePlant Race = "Plant"
	// RacePsychic is the "Psychic" monster race.
	RacePsychic Race = "Psychic"
	// RacePyro is the "Pyro" monster race.
	RacePyro Race = "Pyro"
	// RaceReptile is the "Reptile" monster race.
	RaceReptile Race = "Reptile"
	// RaceRock is the "Rock" monster race.
	RaceRock Race = "Rock"
	// RaceSeaSerpent is the "Sea Serpent" monster race.
	RaceSeaSerpent Race = "Sea Serpent"
	// RaceSpellcaster is the "Spellcaster" monster race.
	RaceSpellcaster Race = "Spellcaster"
	// RaceThunder is the "Thunder" monster race.
	RaceThunder Race = "Thunder"
	// RaceWarrior is the "Warrior" monster race.
	RaceWarrior Race = "Warrior"
	// RaceWingedBeast is the "Winged Beast" monster race.
	RaceWingedBeast Race = "Winged Beast"
	// RaceNormal is the race value for a normal magic or trap card.
	RaceNormal Race = "Normal"
	// RaceField is the race value for a field magic card.
	RaceField Race = "Field"
	// RaceEquip is the race value for an equip magic card.
	RaceEquip Race = "Equip"
	// RaceContinuous is the race value for a continuous magic or trap card.
	RaceContinuous Race = "Continuous"
	// RaceQuickPlay is the race value for a quick-play magic card.
	RaceQuickPlay Race = "Quick-Play"
	// RaceRitual is the race value for a ritual magic card.
	RaceRitual Race = "Ritual"
	// RaceCounter is the race value for a continuous magic or trap card.
	RaceCounter Race = "Counter"
)

// Type of a card.
type Type string

// IsMonster returns whether or not a card is a monster.
func (t Type) IsMonster() bool {
	return strings.HasSuffix(string(t), " Monster")
}

// IsXYZ returns whether or not a card is a XYZ monster.
func (t Type) IsXYZ() bool {
	return strings.HasPrefix(string(t), "XYZ ")
}

// IsSpell returns whether or not a card is a spell.
func (t Type) IsSpell() bool {
	return t == TypeSpellCard
}

// IsTrap returns whether or not a card is a trap.
func (t Type) IsTrap() bool {
	return t == TypeTrapCard
}

// IsSkill returns whether or not a card is a skill (from Duel Links).
func (t Type) IsSkill() bool {
	return t == TypeSkillCard
}

const (
	// TypeEffectMonster is the "Effect Monster" card type.
	TypeEffectMonster Type = "Effect Monster"
	// TypeFlipEffectMonster is the "Flip Effect Monster" card type.
	TypeFlipEffectMonster Type = "Flip Effect Monster"
	// TypeFlipTunerEffectMonster is the "Flip Tuner Effect Monster" card type.
	TypeFlipTunerEffectMonster Type = "Flip Tuner Effect Monster"
	// TypeGeminiMonster is the "Gemini Monster" card type.
	TypeGeminiMonster Type = "Gemini Monster"
	// TypeNormalMonster is the "Normal Monster" card type.
	TypeNormalMonster Type = "Normal Monster"
	// TypeNormalTunerMonster is the "Normal Tuner Monster" card type.
	TypeNormalTunerMonster Type = "Normal Tuner Monster"
	// TypePendulumEffectFusionMonster is the "Pendulum Effect Fusion Monster" card type.
	TypePendulumEffectFusionMonster Type = "Pendulum Effect Fusion Monster"
	// TypePendulumEffectMonster is the "Pendulum Effect Monster" card type.
	TypePendulumEffectMonster Type = "Pendulum Effect Monster"
	// TypePendulumFlipEffectMonster is the "Pendulum Flip Effect Monster" card type.
	TypePendulumFlipEffectMonster Type = "Pendulum Flip Effect Monster"
	// TypePendulumNormalMonster is the "Pendulum Normal Monster" card type.
	TypePendulumNormalMonster Type = "Pendulum Normal Monster"
	// TypePendulumTunerEffectMonster is the "Pendulum Tuner Effect Monster" card type.
	TypePendulumTunerEffectMonster Type = "Pendulum Tuner Effect Monster"
	// TypeRitualEffectMonster is the "Ritual Effect Monster" card type.
	TypeRitualEffectMonster Type = "Ritual Effect Monster"
	// TypeRitualMonster is the "Ritual Monster" card type.
	TypeRitualMonster Type = "Ritual Monster"
	// TypeSpiritMonster is the "Spirit Monster" card type.
	TypeSpiritMonster Type = "Spirit Monster"
	// TypeToonMonster is the "Toon Monster" card type.
	TypeToonMonster Type = "Toon Monster"
	// TypeTunerMonster is the "Tuner Monster" card type.
	TypeTunerMonster Type = "Tuner Monster"
	// TypeUnionEffectMonster is the "Union Effect Monster" card type.
	TypeUnionEffectMonster Type = "Union Effect Monster"
	// TypeUnionTunerEffectMonster is the "Union Tuner Effect Monster" card type.
	TypeUnionTunerEffectMonster Type = "Union Tuner Effect Monster"
	// TypeFusionMonster is the "Fusion Monster" card type.
	TypeFusionMonster Type = "Fusion Monster"
	// TypeLinkMonster is the "Link Monster" card type.
	TypeLinkMonster Type = "Link Monster"
	// TypeSynchroMonster is the "Synchro Monster" card type.
	TypeSynchroMonster Type = "Synchro Monster"
	// TypeSynchroPendulumEffectMonster is the "Synchro Pendulum Effect Monster" card type.
	TypeSynchroPendulumEffectMonster Type = "Synchro Pendulum Effect Monster"
	// TypeSynchroTunerMonster is the "Synchro Tuner Monster" card type.
	TypeSynchroTunerMonster Type = "Synchro Tuner Monster"
	// TypeXYZMonster is the "XYZ Monster" card type.
	TypeXYZMonster Type = "XYZ Monster"
	// TypeXYZPendulumEffectMonster is the "XYZ Pendulum Effect Monster" card type.
	TypeXYZPendulumEffectMonster Type = "XYZ Pendulum Effect Monster"
	// TypeSpellCard is the "Spell Card" card type.
	TypeSpellCard Type = "Spell Card"
	// TypeTrapCard is the "Trap Card" card type.
	TypeTrapCard Type = "Trap Card"
	// TypeSkillCard is the "Skill Card" card type.
	TypeSkillCard Type = "Skill Card"
)

// BanStatus is the tournament ban status of a card.
type BanStatus string

const (
	// BanStatusBanned represents a Banned (also called Forbidden or 禁止)
	// status.
	BanStatusBanned BanStatus = "Banned"
	// BanStatusLimited represents a Limited ban status (制限).
	// Only one copy of the card is allowed in the main, side and extra
	// decks combined.
	BanStatusLimited BanStatus = "Limited"
	// BanStatusSemiLimited represents a Semi-Limited ban status (準制限).
	// Only two copies of the card are allowed in the main, side and extra
	// decks combined.
	BanStatusSemiLimited BanStatus = "Semi-Limited"
)

// LinkMarker is a link marker found on a Link card.
type LinkMarker string

const (
	// LinkMarkerTopLeft is a link marker found on the top-left of a Link card.
	LinkMarkerTopLeft LinkMarker = "Top-Left"
	// LinkMarkerTop is a link marker found on the top of a Link card.
	LinkMarkerTop LinkMarker = "Top"
	// LinkMarkerTopRight is a link marker found on the top-right of a Link card.
	LinkMarkerTopRight LinkMarker = "Top-Right"
	// LinkMarkerRight is a link marker found on the right of a Link card.
	LinkMarkerRight LinkMarker = "Right"
	// LinkMarkerBottomRight is a link marker found on the bottom-right of a
	// Link card.
	LinkMarkerBottomRight LinkMarker = "Bottom-Right"
	// LinkMarkerBottom is a link marker found on the bottom of a Link card.
	LinkMarkerBottom LinkMarker = "Bottom"
	// LinkMarkerBottomLeft is a link marker found on the bottom-left of a
	// Link card.
	LinkMarkerBottomLeft LinkMarker = "Bottom-Left"
	// LinkMarkerLeft is a link marker found on the left of a Link card.
	LinkMarkerLeft LinkMarker = "Left"
)

// CardSet represents the set in which a card is found.
type CardSet struct {
	// Name is the set name.
	Name string `json:"set_name"`
	// Name is the set code.
	Code string `json:"set_code"`
	// Rarity is the card rarity.
	Rarity string `json:"set_rarity"`
	// Price is the price of the card.
	Price string `json:"set_price"`
}

// BanListInfo represents the ban status of the card in several tournament
// formats.
type BanListInfo struct {
	// BanTCG represents the ban status of the card in the TCG.
	BanTCG *BanStatus `json:"ban_tcg"`
	// BanOCG represents the ban status of the card in the OCG.
	BanOCG *BanStatus `json:"ban_ocg"`
	// BanGOAT represents the ban status of the card in the Goat format.
	BanGOAT *BanStatus `json:"ban_goat"`
}

// CardImage represents the image of a card.
type CardImage struct {
	// ID is the card image ID.
	ID int64 `json:"id"`
	// URL is the card URL.
	URL string `json:"image_url"`
	// URLSmall is the card thumbnail URL.
	URLSmall string `json:"image_url_small"`
}

// CardPrice represents the price information of a card.
type CardPrice struct {
	// CardMarketPrice represents the price of a card on
	// https://www.cardmarket.com/.
	CardMarketPrice string `json:"cardmarket_price"`
	// TCGPlayerPrice represents the price of a card on
	// https://www.tcgplayer.com/.
	TCGPlayerPrice string `json:"tcgplayer_price"`
	// CoolStuffIncPrice represents the price of a card on
	// https://www.coolstuffinc.com/.
	CoolStuffIncPrice string `json:"coolstuffinc_price"`
	// EbayPrice represents the price of a card on Ebay.
	EbayPrice string `json:"ebay_price"`
	// AmazonPrice represents the price of a card on Amazon.
	AmazonPrice string `json:"amazon_price"`
}

// Data is the YGOProDeck API response struct.
type Data struct {
	YGOProID    int64        `json:"id"`
	Name        string       `json:"name"`
	Type        Type         `json:"type"`
	Description string       `json:"desc"`
	Attack      *int         `json:"atk"`
	Defense     *int         `json:"def"`
	Level       *int         `json:"level"`
	Race        Race         `json:"race"`
	Attribute   *Attribute   `json:"attribute"`
	Scale       *int         `json:"scale"`
	LinkValue   *int         `json:"linkval"`
	LinkMarkers []LinkMarker `json:"linkmarkers"`
	Archetype   *string      `json:"archetype"`
	Sets        []CardSet    `json:"card_sets"`
	BanListInfo *BanListInfo `json:"banlist_info"`
	Images      []CardImage  `json:"card_images"`
	Prices      []CardPrice  `json:"card_prices"`
}

type clientOptions struct {
	baseURL string
	client  *http.Client
}

// ClientOption configures the API client.
type ClientOption func(*clientOptions)

// WithBaseURL returns an option which overrides the base URL.
func WithBaseURL(baseURL string) ClientOption {
	return func(o *clientOptions) {
		o.baseURL = baseURL
	}
}

// WithHTTPClient returns an option which overrides the default HTTP client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(o *clientOptions) {
		o.client = client
	}
}

// QueryName sends a request to the YGOProDeck API to retrieve data about a card from its name.
func QueryName(name string, options ...ClientOption) (data Data, err error) {
	return query("name", name, options...)
}

// QueryID sends a request to the YGOProDeck API to retrieve data about a card from its YGOProDeck ID.
func QueryID(id int64, options ...ClientOption) (data Data, err error) {
	return query("id", strconv.FormatInt(id, 10), options...)
}

func query(paramName string, paramValue string, options ...ClientOption) (data Data, err error) {
	// Default options
	co := &clientOptions{
		baseURL: defaultBaseURL,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
	for _, option := range options {
		option(co)
	}

	// Parse the URL and add "?name={name}" to it
	url, err := url.Parse(co.baseURL)
	if err != nil {
		return
	}
	query := url.Query()
	query.Set(paramName, paramValue)
	url.RawQuery = query.Encode()

	targetURL := url.String()

	// Build the request
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return
	}

	// Send the request
	resp, err := co.client.Do(req)
	if err != nil {
		return
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			err = closeErr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("received invalid status code %d", resp.StatusCode)
		return
	}

	// Fill the record with the data from the JSON
	var record []Data

	// Use json.Decode for reading streams of JSON data
	err = json.NewDecoder(resp.Body).Decode(&record)
	if err != nil {
		return
	}

	if len(record) == 0 {
		err = errors.New("received an empty response")
		return
	}

	if len(record[0].Images) == 0 {
		err = errors.New("no image associated to card")
	}

	// Even if we received multiple responses, return only the first one
	data = record[0]

	return
}
