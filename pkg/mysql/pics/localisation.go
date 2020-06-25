package pics

type Localisation struct {
	RichPresence map[string]*LocalisationLanguage `json:"richpresence"`
}

func (l *Localisation) AddLanguage(key string, val *LocalisationLanguage) {

	if l.RichPresence == nil {
		l.RichPresence = map[string]*LocalisationLanguage{}
	}

	l.RichPresence[key] = val
}

func (l Localisation) HasLanguages() bool {
	return len(l.RichPresence) > 0
}

type LocalisationLanguage struct {
	Tokens LocalisationTokens `json:"tokens"`
}

func (l *LocalisationLanguage) AddToken(key string, val string) {

	if l.Tokens == nil {
		l.Tokens = LocalisationTokens{}
	}

	l.Tokens[key] = &val
}

type LocalisationTokens map[string]*string
