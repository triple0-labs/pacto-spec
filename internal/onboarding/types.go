package onboarding

type Intents struct {
	Problem string
}

type Sources struct {
	Languages string
	Tools     string
}

type Profile struct {
	Languages       []string
	CustomLanguages []string
	Tools           []string
	CustomTools     []string
	Intents         Intents
	Sources         Sources
}
