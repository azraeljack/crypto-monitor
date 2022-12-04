package binance

type Config struct {
	ApiKey    string `json:"api_key"`
	ApiSecret string `json:"api_secret"`
	Timeout   string `json:"timeout"`
	Interval  string `json:"interval"`
	Proxy     string `json:"proxy"`
}
