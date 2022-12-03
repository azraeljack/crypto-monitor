package price_change

type Config struct {
	Symbol1    string  `json:"symbol1"`
	Symbol2    string  `json:"symbol2"`
	WindowSize string  `json:"window_size"`
	Absolute   float64 `json:"absolute"`
	Percentage float64 `json:"percentage"`
}
