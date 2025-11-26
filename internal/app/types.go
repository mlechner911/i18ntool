package app

type TranslationManager struct {
	files     map[string]string
	data      map[string]map[string]interface{}
	Languages []string
}

type MissingTranslation struct {
	Key          string
	Translations map[string]string
}
