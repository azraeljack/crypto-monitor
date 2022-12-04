package wechat

type Config struct {
	Timeout    string `json:"timeout"`
	WebhookURL string `json:"webhook_url"`
	Throttle   string `json:"throttle"`
}
