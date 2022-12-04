package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/azraeljack/crypto-monitor/notifier"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

type TextContent struct {
	Content string `json:"content"`
}

type NotificationMsg struct {
	MsgType string      `json:"msgtype"`
	Text    TextContent `json:"text"`
}

func (n *NotificationMsg) ToJSON() []byte {
	raw, _ := json.Marshal(n)
	return raw
}

type Notifier struct {
	webhookURL string
	httpClient *http.Client

	throttle time.Duration

	lastNotifiedTime atomic.Int64
	ctx              context.Context
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

	throttle, err := time.ParseDuration(conf.Throttle)
	if err != nil {
		throttle = 5 * time.Second
	}

	return &Notifier{
		webhookURL: conf.WebhookURL,
		throttle:   throttle,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		ctx: ctx,
	}
}

func (w *Notifier) Notify(msg string, throttle bool) {
	log.Info("sending wechat notification...")
	log.Debugf("wechat payload: %v", msg)

	if throttle {
		currentTime := time.Now().UnixNano()
		if w.lastNotifiedTime.Load()+int64(w.throttle) > currentTime {
			log.Infof("wechat notifiation throttled, ignore message")
			return
		}

		w.lastNotifiedTime.Store(currentTime)
	}

	payload := &NotificationMsg{
		MsgType: "text",
		Text: TextContent{
			Content: msg,
		},
	}
	request, err := http.NewRequestWithContext(w.ctx, http.MethodPost, w.webhookURL, bytes.NewReader(payload.ToJSON()))
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
