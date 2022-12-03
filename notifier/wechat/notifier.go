package wechat

import (
	"context"
	"encoding/json"
	"github.com/azraeljack/crypto-monitor/notifier"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

type Notifier struct {
	webhookURL string
	httpClient *http.Client

	ctx context.Context
}

func NewWechatNotifier(ctx context.Context, rawConf json.RawMessage) notifier.Notifier {
	conf := &Config{}
	if err := json.Unmarshal(rawConf, conf); err != nil {
		log.Panic("failed to parse wechat notifier config", err)
	}

	timeout, err := time.ParseDuration(conf.Timeout)
	if err != nil {
		timeout = 5 * time.Second
	}

	return &Notifier{
		webhookURL: conf.WebhookURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		ctx: ctx,
	}
}

func (w *Notifier) Notify(msg string) {
	log.Info("sending wechat notification...")
	log.Debugf("wechat payload: %v", msg)

	request, err := http.NewRequestWithContext(w.ctx, http.MethodPost, w.webhookURL, strings.NewReader(msg))
	if err != nil {
		log.Errorf("notify fail: %v", err)
		return
	}
	request.Header.Add("content-type", "application/json")

	resp, err := w.httpClient.Do(request)
	if err != nil {
		log.Errorf("notify fail: %v", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		errorMsg, _ := io.ReadAll(resp.Body)
		log.Errorf("notify fail: %s", errorMsg)
	}
	log.Info("successfully notified via wechat")
}