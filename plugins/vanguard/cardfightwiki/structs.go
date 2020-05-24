package cardfightwiki

// Card contains the data found on a Vanguartd card page from
// https://cardfight.fandom.com/
type Card struct {
	EnglishName      string
	JapaneseName     string
	Type             *string
	Grade            int
	Skill            *string
	Power            *string
	Critical         *int
	Shield           *int
	Nation           *string
	Clan             *string
	Race             *string
	TriggerEffect    *string
	Formats          []string
	Flavor           *string
	Effect           *string
	EnglishImageURL  string
	JapaneseImageURL string
}
