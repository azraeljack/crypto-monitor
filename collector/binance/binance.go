package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/azraeljack/crypto-monitor/collector"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

type Collector struct {
	config *Config

	ctx context.Context

	client   *binance.Client
	timeout  time.Duration
	interval time.Duration
}

func (c *Collector) TestConnection() bool {
	return c.client.NewPingService().Do(c.getContext()) == nil
}

func (c *Collector) CollectWindowPrice(symbol1, symbol2 string, window time.Duration) <-chan *collector.WindowPrice {
	resultCh := make(chan *collector.WindowPrice, 20)

	go func() {
		pair := combineSymbols(symbol1, symbol2)
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-c.ctx.Done():
				close(resultCh)
				log.Info("binance collector exited")
				return
			case <-ticker.C:
				log.Infof("sending new window price request of [%s - %s] to binance...", symbol1, symbol2)
				res, err := c.client.NewListSymbolTickerService().Symbol(pair).WindowSize(fmt.Sprintf("%vm", uint64(window.Minutes()))).Do(c.getContext())
				if err != nil {
					log.Errorf("failed to fetch average price_change of [%s-%s], err: %v", symbol1, symbol2, err)
					continue
				} else if len(res) < 1 {
					log.Warnf("failed to fetch average price_change of %s-%s, result empty", symbol1, symbol2)
					continue
				}
				price := res[0]
				log.Debugf("received response from binance %s", toJSONString(res))

				windowPrice := &collector.WindowPrice{
					Symbol1:             symbol1,
					Symbol2:             symbol2,
					OpenPrice:           stringToFloat(price.OpenPrice),
					ClosePrice:          stringToFloat(price.LastPrice),
					HighPrice:           stringToFloat(price.HighPrice),
					LowPrice:            stringToFloat(price.LowPrice),
					Volume:              stringToFloat(price.Volume),
					QuoteVolume:         stringToFloat(price.QuoteVolume),
					AbsolutePriceChange: stringToFloat(price.PriceChange),
					RelativePriceChange: stringToFloat(price.PriceChangePercent),
					OpenTime:            uint64(price.OpenTime),
					CloseTime:           uint64(price.CloseTime),
					OrderCount:          uint64(price.Count),
				}

				select {
				case resultCh <- windowPrice:
					log.Debugf("fetched new window price_change of [%s-%s]: %s", symbol1, symbol2, windowPrice.String())
				default:
					log.Warnf("result channel full of [%s-%s], discard data: %s", symbol1, symbol2, windowPrice.String())
				}
			}
		}
	}()

	return resultCh
}

func (c *Collector) CollectAvgPrice(symbol1, symbol2 string) <-chan float64 {
	resultCh := make(chan float64, 20)

	go func() {
		pair := combineSymbols(symbol1, symbol2)
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-c.ctx.Done():
				close(resultCh)
				log.Info("collector exited")
			case <-ticker.C:
				res, err := c.client.NewAveragePriceService().Symbol(pair).Do(c.getContext())
				if err != nil {
					log.Errorf("failed to fetch average price_change of %s-%s, err: %v", symbol1, symbol2, err)
					continue
				}

				price := stringToFloat(res.Price)
				if price == 0.0 {
					continue
				}

				select {
				case resultCh <- price:
					log.Debugf("fetched new price_change of %s-%s: %v", symbol1, symbol2, price)
				default:
					log.Warnf("result channel full of %s-%s", symbol1, symbol2)
				}
			}
		}
	}()

	return resultCh
}

func (c *Collector) Type() string {
	return "binance"
}

func (c *Collector) getContext() context.Context {
	ctx, _ := context.WithTimeout(c.ctx, c.timeout)
	return ctx
}

func combineSymbols(symbols ...string) string {
	builder := &strings.Builder{}
	for _, s := range symbols {
		builder.WriteString(strings.ToUpper(s))
	}

	return builder.String()
}

func stringToFloat(str string) float64 {
	n, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.Errorf("failed to parse %s", str)
		return 0.0
	}
	return n
}

func toJSONString(data any) string {
	res, _ := json.Marshal(data)
	return string(res)
}

func NewBinanceCollector(ctx context.Context, rawConf json.RawMessage) collector.Collector {
	conf := &Config{}
	if err := json.Unmarshal(rawConf, conf); err != nil {
		log.Panic("failed to parse binance collector config", err)
	}

	timeout, err := time.ParseDuration(conf.Timeout)
	if err != nil {
		timeout = 5 * time.Second
	}

	interval, err := time.ParseDuration(conf.Interval)
	if err != nil {
		interval = 5 * time.Second
	}

	var client *binance.Client
	if len(conf.Proxy) > 0 {
		client = binance.NewProxiedClient(conf.ApiKey, conf.ApiSecret, conf.Proxy)
	} else {
		client = binance.NewClient(conf.ApiKey, conf.ApiSecret)
	}

	return &Collector{
		config:   conf,
		client:   client,
		timeout:  timeout,
		interval: interval,
		ctx:      ctx,
	}
}
