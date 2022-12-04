package price_change

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/azraeljack/crypto-monitor/collector"
	"github.com/azraeljack/crypto-monitor/notifier"
	"github.com/azraeljack/crypto-monitor/strategy"
	cache "github.com/go-pkgz/expirable-cache/v2"
	log "github.com/sirupsen/logrus"
	"html/template"
	"math"
	"time"
)

var notificationTemplate = `发现价格波动：
- 时间：{{.Time}}
- 交易对：{{.Symbol1}} - {{.Symbol2}}
- 波动幅度：{{.Absolute}} ({{.Percentage}}%)
- 成交笔数：{{.OrderNumber}}
`

type Notification struct {
	Time        string
	Symbol1     string
	Symbol2     string
	Absolute    string
	Percentage  string
	OrderNumber string
}

func NewNotification(symbol1, symbol2 string, price *collector.WindowPrice) *Notification {
	return &Notification{
		Time:        time.Now().Format("2006-01-02 15:04:05"),
		Symbol1:     symbol1,
		Symbol2:     symbol2,
		Absolute:    fmt.Sprintf("%v", price.AbsolutePriceChange),
		Percentage:  fmt.Sprintf("%v", price.RelativePriceChange),
		OrderNumber: fmt.Sprintf("%v", price.OrderCount),
	}
}

type Strategy struct {
	windowSize time.Duration

	symbol1 string
	symbol2 string

	absolute   float64
	percentage float64

	ctx        context.Context
	priceCache cache.Cache[string, struct{}]

	collectors []collector.Collector
	notifiers  []notifier.Notifier
}

func (s *Strategy) AddCollectors(collector ...collector.Collector) {
	s.collectors = append(s.collectors, collector...)
}

func (s *Strategy) AddNotifiers(notifier ...notifier.Notifier) {
	s.notifiers = append(s.notifiers, notifier...)
}

func (s *Strategy) Run() {
	log.Infof("start running price change strategy for [%s - %s]", s.symbol1, s.symbol2)
	notifyPriceCh := make(chan *collector.WindowPrice, len(s.collectors)*20+1)

	for _, c := range s.collectors {
		go func(col collector.Collector) {
			newPrice := col.CollectWindowPrice(s.symbol1, s.symbol2, s.windowSize)
			for {
				select {
				case price, ok := <-newPrice:
					if !ok {
						log.Info("price change strategy collector listener exit")
						return
					}
					priceKey := fmt.Sprintf("%v", price.AbsolutePriceChange)
					if _, exist := s.priceCache.Peek(priceKey); exist {
						log.Infof("price already notified in this window")
						continue
					}
					if math.Abs(price.AbsolutePriceChange) >= s.absolute || math.Abs(price.RelativePriceChange) >= s.percentage {
						s.priceCache.Set(priceKey, struct{}{}, s.windowSize)

						select {
						case notifyPriceCh <- price:
							log.Infof("received strategy matched price change [%s - %s]: %s", s.symbol1, s.symbol2, price.String())
						default:
							log.Warnf("price change notify channel full, discard data: %s", price.String())
						}
					} else {
						log.Debugf("received unmatched price change, absolute: %v, relative: %v", price.AbsolutePriceChange, price.RelativePriceChange)
					}
				case <-s.ctx.Done():
					log.Info("price change strategy collector listener exit")
					return
				}
			}
		}(c)
	}

	go func() {
		for {
			select {
			case price := <-notifyPriceCh:
				for _, n := range s.notifiers {
					go func(price *collector.WindowPrice, not notifier.Notifier) {
						log.Info("sending price change notification...")
						log.Debugf("price change: %v", price.String())
						tmpl := template.New("PriceChangeNotification")
						if _, err := tmpl.Parse(notificationTemplate); err != nil {
							log.Warnf("unable to parse template: %v", err)
							return
						}

						stringWriter := bytes.NewBufferString("")
						if err := tmpl.Execute(stringWriter, NewNotification(s.symbol1, s.symbol2, price)); err != nil {
							log.Warnf("unable to redner template: %v", err)
							return
						}

						not.Notify(stringWriter.String(), "PriceChange^"+price.SymbolPair(), true)
						log.Infof("price change notifcation sent")
					}(price, n)
				}
			case <-s.ctx.Done():
				log.Infof("price change notifer worker exit")
				return
			}
		}
	}()
}

func NewPriceChangeStrategy(ctx context.Context, rawConf json.RawMessage) strategy.Strategy {
	conf := &Config{}
	if err := json.Unmarshal(rawConf, conf); err != nil {
		log.Panic("failed to parse price change strategy config", err)
	}

	windowSize, err := time.ParseDuration(conf.WindowSize)
	if err != nil {
		windowSize = 15 * time.Minute
	}

	return &Strategy{
		windowSize: windowSize,
		symbol1:    conf.Symbol1,
		symbol2:    conf.Symbol2,
		absolute:   conf.Absolute,
		percentage: conf.Percentage,
		ctx:        ctx,
		priceCache: cache.NewCache[string, struct{}]().WithTTL(windowSize),
		collectors: make([]collector.Collector, 0),
		notifiers:  make([]notifier.Notifier, 0),
	}
}
