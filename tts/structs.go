package tts

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// Decimal is a float64 that is serialized in JSON with a trailing 0 if it doesn't have a decimal part.
type Decimal float64

// MarshalJSON implements the json.Marshaler interface.
// If the value is an integer, add a trailer ".0" when serializing.
func (d Decimal) MarshalJSON() ([]byte, error) {
	f := float64(d)

	if math.IsInf(f, 0) || math.IsNaN(f) {
		return nil, errors.New("unsupported value")
	}

	str := strconv.FormatFloat(f, 'f', -1, 32)

	if !strings.Contains(str, ".") {
		// Add a trailing 0 if it's not a decimal number
		str += ".0"
	}

	return []byte(str), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Decimal) UnmarshalJSON(text []byte) error {
	t := string(text)
	if t == "null" {
		return nil
	}
	i, err := strconv.ParseFloat(t, 64)
	if err != nil {
		return err
	}
	*d = Decimal(i)
	return nil
}

// ObjectType is the type of a TTS object.
type ObjectType string

const (
	// DeckObject represents a card deck.
	DeckObject ObjectType = "Deck"
	// DeckCustomObject represents a custom card deck.
	DeckCustomObject ObjectType = "DeckCustom"
	// CardObject represents a card.
	CardObject ObjectType = "Card"
	// CardCustomObject represents a custom card.
	CardCustomObject ObjectType = "CardCustom"
)

// DefaultTransform is the object transform data used by default in TTS.
var DefaultTransform Transform = Transform{
	PosX:   0,
	PosY:   0,
	PosZ:   0,
	RotX:   0,
	RotY:   180,
	RotZ:   180,
	ScaleX: 1,
	ScaleY: 1,
	ScaleZ: 1,
}

// DefaultColorDiffuse is the color diffuse data used by default in TTS.
var DefaultColorDiffuse ColorDiffuse = ColorDiffuse{
	Red:   0.713239133,
	Green: 0.713239133,
	Blue:  0.713239133,
}

// SavedObject represents an object saved in the TTS chest
// (also used for save files).
// See https://kb.tabletopsimulator.com/custom-content/save-file-format/.
type SavedObject struct {
	// SaveName is the name of the saved object.
	SaveName string `json:"SaveName"`
	// GameMode
	GameMode string `json:"GameMode"`
	// Gravity
	Gravity Decimal `json:"Gravity"`
	// PlayArea
	PlayArea Decimal `json:"PlayArea"`
	// Date
	Date string `json:"Date"`
	// Table
	Table string `json:"Table"`
	// Sky, for custom sky
	Sky string `json:"Sky"`
	// Note
	Note string `json:"Note"`
	// Rules
	Rules string `json:"Rules"`
	// LuaScript contains a custom Lua script.
	LuaScript string `json:"LuaScript"`
	// LuaScript contains the state of the custom Lua script.
	LuaScriptState string `json:"LuaScriptState"`
	// XMLUI contains a custom XML UI.
	XMLUI string `json:"XmlUI"`
	// ObjectStates contains the objects on the table.
	ObjectStates []Object `json:"ObjectStates"`
	// TabStates contains the notepad tabs.
	TabStates struct{} `json:"TabStates"`
	// VersionNumber is the version number of the save state.
	VersionNumber string `json:"VersionNumber"`
}

// CustomDeckMap is a map of CustomDeck, whose keys are the custom deck indexes.
type CustomDeckMap map[string]CustomDeck

// MarshalJSON implements the json.Marshaler interface.
// The keys are strings, but contain integers only. They are serialized in
// order as if they were integer instead of strings (i.e. "2" will be serialized
// after "1", instead of "10").
func (cdm CustomDeckMap) MarshalJSON() ([]byte, error) {
	length := len(cdm)

	// Convert the keys to integer and sort them in a slice
	keys := make([]int, 0, length)
	for key := range cdm {
		intKey, err := strconv.Atoi(key)
		if err != nil {
			// The key cannot be converted to an integer
			// This shouldn't happen
			return nil, err
		}
		keys = append(keys, intKey)
	}

	// Sort the keys, in integer order
	sort.Ints(keys)

	// Serialize the map
	buffer := bytes.NewBufferString("{")
	count := 0

	// Iterate through the ordered keys
	for _, key := range keys {
		strKey := strconv.Itoa(key)
		value, ok := cdm[strKey]
		if !ok {
			return nil, errors.New("key " + strKey + " not found")
		}
		jsonValue, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		buffer.WriteString(fmt.Sprintf("\"%d\":%s", key, string(jsonValue)))
		count++
		if count < length {
			buffer.WriteString(",")
		}
	}

	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

// Object is a TTS object.
type Object struct {
	// ObjectType represents the type of the object.
	ObjectType ObjectType `json:"Name"`
	// Transform contains the position, rotation and scale data of the object.
	Transform Transform `json:"Transform"`
	// Nickname is the object name.
	Nickname string `json:"Nickname"`
	// Description of the object.
	Description string `json:"Description"`
	// GM notes attached to the object.
	GMNotes string `json:"GMNotes"`
	// ColorDiffuse is the color information of the object.
	ColorDiffuse ColorDiffuse `json:"ColorDiffuse"`
	// Locked, when set, freezes an object in place, stopping all physical
	// interactions.
	Locked bool `json:"Locked"`
	// Grid makes the object snap to any grid point.
	Grid bool `json:"Grid"`
	// Snap makes the object snap to any snap point.
	Snap bool `json:"Snap"`
	// IgnoreFoW makes the object visible even inside fog of war.
	IgnoreFoW bool `json:"IgnoreFoW"`
	// MeasureMovement makes the measure tool be automatically used when moving this object.
	MeasureMovement bool `json:"MeasureMovement"`
	// DragSelectable makes an object be selected in a drag selection.
	DragSelectable bool `json:"DragSelectable"`
	// Autoraise makes the object automatically raise above potential collisions.
	Autoraise bool `json:"Autoraise"`
	// Sticky makes the objects above this one attached to it when it is picked
	// up.
	Sticky bool `json:"Sticky"`
	// Show a tooltip when hovering over the object (name, description, icon).
	Tooltip bool `json:"Tooltip"`
	// Should this object receive grid lines projected onto it?
	GridProjection bool `json:"GridProjection"`
	// When object is face down, it will be hidden as a question mark.
	HideWhenFaceDown bool `json:"HideWhenFaceDown"`
	// Should this object go into the players' hand?
	Hands bool `json:"Hands"`
	// CardID is the ID of the card in the deck.
	CardID int `json:"CardID,omitempty"`
	// SidewaysCard is whether or not the card should be displayed sideways.
	SidewaysCard bool `json:"SidewaysCard"`
	// DeckIDs are of IDs of the card found in the deck.
	DeckIDs []int `json:"DeckIDs,omitempty"`
	// CustomDeck contains the information of the cards in the deck.
	CustomDeck CustomDeckMap `json:"CustomDeck,omitempty"`
	// XMLUI contains a custom XML UI.
	XMLUI string `json:"XmlUI"`
	// LuaScript contains a custom Lua script.
	LuaScript string `json:"LuaScript"`
	// LuaScript contains the state of the custom Lua script.
	LuaScriptState string `json:"LuaScriptState"`
	// ContainedObjects represents the objects contained by this object.
	ContainedObjects []Object `json:"ContainedObjects,omitempty"`
	// States lists the differents states of the object.
	// See https://berserk-games.com/knowledgebase/creating-states/.
	States map[string]Object `json:"States,omitempty"`
	// GUID is the Globally Unique Identifier of the object.
	GUID string `json:"GUID"`
}

// Transform contains the position, rotation and scale data of an object.
type Transform struct {
	// PosX is the X position of the object.
	PosX Decimal `json:"posX"`
	// PosY is the Y position of the object.
	PosY Decimal `json:"posY"`
	// PosZ is the Z position of the object.
	PosZ Decimal `json:"posZ"`
	// RotX is the rotation on the X-axis.
	RotX Decimal `json:"rotX"`
	// RotY is the rotation on the Y-axis.
	RotY Decimal `json:"rotY"`
	// RotZ is the rotation on the Z-axis.
	RotZ Decimal `json:"rotZ"`
	// ScaleX is the scale on the X-axis.
	ScaleX Decimal `json:"scaleX"`
	// ScaleY is the scale on the Y-axis.
	ScaleY Decimal `json:"scaleY"`
	// ScaleZ is the scale on the Z-axis.
	ScaleZ Decimal `json:"scaleZ"`
}

// ColorDiffuse is the color information of an object.
type ColorDiffuse struct {
	// Red color diffuse.
	Red Decimal `json:"r"`
	// Green color diffuse.
	Green Decimal `json:"g"`
	// Blue color diffuse.
	Blue Decimal `json:"b"`
}

// DeckShape is the shape of the custom deck.
type DeckShape int

const (
	// DeckShapeRectangleRounded is the default deck shape.
	DeckShapeRectangleRounded DeckShape = iota
	// DeckShapeRectangle is the rectangle deck shape.
	DeckShapeRectangle DeckShape = iota
	// DeckShapeHexRounded is the hex (rounded) deck shape.
	DeckShapeHexRounded DeckShape = iota
	// DeckShapeHex is the hex deck shape.
	DeckShapeHex DeckShape = iota
	// DeckShapeCircle is the circle deck shape.
	DeckShapeCircle DeckShape = iota
)

// CustomDeck represents a custom TTS deck.
// See https://berserk-games.com/knowledgebase/custom-decks/.
type CustomDeck struct {
	// FaceURL is the address of the card faces.
	FaceURL string `json:"FaceURL"`
	// BackURL is the address of the card back (backs if UniqueBack is true).
	BackURL string `json:"BackURL"`
	// NumWidth is the number of cards in a single row of the face image
	// (and back image if UniqueBack is true).
	NumWidth int `json:"NumWidth"`
	// NumHeight is the number of cards in a single column of the face image
	// (and back image if UniqueBack is true).
	NumHeight int `json:"NumHeight"`
	// BackIsHidden determines if the BackURL should be used as the back of the
	// cards instead of the last image of the card face image.
	BackIsHidden bool `json:"BackIsHidden"`
	// UniqueBack should be true if each card is using a different back.
	UniqueBack bool `json:"UniqueBack"`
	// Type is the shape of the deck.
	Type DeckShape `json:"Type"`
}

func createSavedObject(objectStates []Object) SavedObject {
	return SavedObject{
		Gravity:      0.5,
		PlayArea:     0.5,
		ObjectStates: objectStates,
	}
}

func createDefaultDeck() SavedObject {
	return createSavedObject([]Object{
		{
			// TODO: Find the difference between "Deck" and "DeckCustom"
			// The Scryfall mod uses "Deck" while Decker uses "DeckCustom"
			// ObjectType:       DeckCustomObject,
			ObjectType:       DeckObject,
			Transform:        DefaultTransform,
			ColorDiffuse:     DefaultColorDiffuse,
			Locked:           false,
			Grid:             true,
			Snap:             true,
			IgnoreFoW:        false,
			MeasureMovement:  false,
			DragSelectable:   true,
			Autoraise:        true,
			Sticky:           true,
			Tooltip:          true,
			GridProjection:   false,
			HideWhenFaceDown: true,
			Hands:            false,
			SidewaysCard:     false,
			CustomDeck:       make(CustomDeckMap),
		},
	})
}
