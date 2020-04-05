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
	AttributeWind Attribute = "WING"
	// AttributeLaugh represents the Laugh (unofficial) monster card attribute.
	AttributeLaugh Attribute = "LAUGH"
)

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

type Type string

func (t Type) IsMonster() bool {
	return strings.HasSuffix(string(t), " Monster")
}

func (t Type) IsXYZ() bool {
	return strings.HasPrefix(string(t), "XYZ ")
}

func (t Type) IsSpell() bool {
	return t == TypeSpellCard
}

func (t Type) IsTrap() bool {
	return t == TypeTrapCard
}

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

type BanStatus string

const (
	BanStatusBanned      BanStatus = "Banned"
	BanStatusLimited     BanStatus = "Limited"
	BanStatusSemiLimited BanStatus = "Semi-Limited"
)

type LinkMarker string

const (
	LinkMarkerTopLeft     LinkMarker = "Top-Left"
	LinkMarkerTop         LinkMarker = "Top"
	LinkMarkerTopRight    LinkMarker = "Top-Right"
	LinkMarkerRight       LinkMarker = "Right"
	LinkMarkerBottomRight LinkMarker = "Bottom-Right"
	LinkMarkerBottom      LinkMarker = "Bottom"
	LinkMarkerBottomLeft  LinkMarker = "Bottom-Left"
	LinkMarkerLeft        LinkMarker = "Left"
)

type CardSet struct {
	Name   string `json:"set_name"`
	Code   string `json:"set_code"`
	Rarity string `json:"set_rarity"`
	Price  string `json:"set_price"`
}

type BanlistInfo struct {
	BanTCG  *BanStatus `json:"ban_tcg"`
	BanOCG  *BanStatus `json:"ban_ocg"`
	BanGOAT *BanStatus `json:"ban_goat"`
}

type CardImage struct {
	ID       int64  `json:"id"`
	URL      string `json:"image_url"`
	URLSmall string `json:"image_url_small"`
}

type CardPrice struct {
	CardMarketPrice   string `json:"cardmarket_price"`
	TCGPlayerPrice    string `json:"tcgplayer_price"`
	CoolStuffIncPrice string `json:"coolstuffinc_price"`
	EbayPrice         string `json:"ebay_price"`
	AmazonPrice       string `json:"amazon_price"`
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
	BanlistInfo *BanlistInfo `json:"banlist_info"`
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
