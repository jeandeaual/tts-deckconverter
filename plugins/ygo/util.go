package ygo

import (
	"strconv"
	"strings"

	"github.com/jeandeaual/tts-deckconverter/plugins/ygo/api"
)

func buildDescription(apiResponse api.Data) string {
	var sb strings.Builder

	if apiResponse.Attribute != nil {
		switch *apiResponse.Attribute {
		case api.AttributeDark:
			sb.WriteString("[000000]")
		case api.AttributeDivine, api.AttributeLight:
			sb.WriteString("[8a7045]")
		case api.AttributeEarth:
			sb.WriteString("[374231]")
		case api.AttributeFire:
			sb.WriteString("[fe0d00]")
		case api.AttributeWater:
			sb.WriteString("[02a2d7]")
		case api.AttributeWind:
			sb.WriteString("[4d8742]")
		case api.AttributeLaugh:
			sb.WriteString("[ee8224]")
		}
		sb.WriteString(string(*apiResponse.Attribute))
		sb.WriteString("[ffffff]\n\n")
	}
	if apiResponse.Level != nil &&
		apiResponse.Type.IsMonster() &&
		apiResponse.Type != api.TypeLinkMonster {
		if apiResponse.Type.IsXYZ() {
			sb.WriteString("[b9b959]Rank ")
		} else {
			sb.WriteString("[ffd33c]Level ")
		}
		sb.WriteString(strconv.Itoa(*apiResponse.Level))
		sb.WriteString("[ffffff]\n\n")
	}
	if apiResponse.Archetype != nil {
		sb.WriteString("[i]")
		sb.WriteString(*apiResponse.Archetype)
		sb.WriteString("[/i]\n\n")
	}
	if apiResponse.Scale != nil {
		sb.WriteString("[2d68dc]Scale [c2243a]")
		sb.WriteString(strconv.Itoa(*apiResponse.Scale))
		sb.WriteString("[ffffff]\n\n")
	}
	if apiResponse.Type.IsMonster() {
		sb.WriteString("[b][ ")
		sb.WriteString(string(apiResponse.Race))
		sb.WriteString(" / ")
		sb.WriteString(
			strings.Replace(
				strings.TrimSuffix(string(apiResponse.Type), " Monster"),
				" ",
				" / ",
				-1,
			),
		)
		sb.WriteString(" ][/b]\n")
	} else {
		sb.WriteString("[b][ ")
		sb.WriteString(string(apiResponse.Type))
		if apiResponse.Race != "Normal" {
			sb.WriteString(" / ")
			sb.WriteString(string(apiResponse.Race))
		}
		sb.WriteString(" ][/b]\n")
	}
	sb.WriteString(strings.Replace(apiResponse.Description, "\r\n", "\n", -1))
	if apiResponse.Attack != nil || apiResponse.Defense != nil {
		sb.WriteString("\n\n")
	}
	if apiResponse.Attack != nil {
		sb.WriteString("[b]ATK/")
		sb.WriteString(strconv.Itoa(*apiResponse.Attack))
		sb.WriteString("[/b] ")
	}
	if apiResponse.Defense != nil {
		sb.WriteString("[b]DEF/")
		sb.WriteString(strconv.Itoa(*apiResponse.Defense))
		sb.WriteString("[/b]")
	}
	if apiResponse.LinkValue != nil {
		sb.WriteString("[b]LINK-")
		sb.WriteString(strconv.Itoa(*apiResponse.LinkValue))
		sb.WriteString("[/b]")
	}
	if apiResponse.LinkMarkers != nil {
		sb.WriteString("\n")
		for i, linkmarker := range apiResponse.LinkMarkers {
			sb.WriteString(string(linkmarker))
			if i < len(apiResponse.LinkMarkers)-1 {
				sb.WriteString(", ")
			}
		}
	}

	return sb.String()
}
