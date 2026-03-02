package onboarding

type Intents struct {
	Problem string
}

type Sources struct {
	Languages string
	Tools     string
	UI        string
}

type Profile struct {
	Languages       []string
	CustomLanguages []string
	Tools           []string
	CustomTools     []string
	UILanguage      string
	Intents         Intents
	Sources         Sources
}
