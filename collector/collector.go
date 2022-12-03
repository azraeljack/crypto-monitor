package collector

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Builder func(ctx context.Context, rawConf json.RawMessage) Collector

type Collector interface {
	CollectAvgPrice(symbol1, symbol2 string) <-chan float64
	CollectWindowPrice(symbol1, symbol2 string, window time.Duration) <-chan *WindowPrice
	Type() string
	TestConnection() bool
}

type WindowPrice struct {
	OpenPrice           float64 `json:"open_price"`
	ClosePrice          float64 `json:"close_price"`
	HighPrice           float64 `json:"high_price"`
	LowPrice            float64 `json:"low_price"`
	Volume              float64 `json:"volume"`
	QuoteVolume         float64 `json:"quote_volume"`
	OpenTime            uint64  `json:"open_time"`
	CloseTime           uint64  `json:"close_time"`
	OrderCount          uint64  `json:"order_count"`
	AbsolutePriceChange float64 `json:"absolute_price_change"`
	RelativePriceChange float64 `json:"relative_price_change"`
}

func (w WindowPrice) String() string {
	s, _ := json.Marshal(w)
	return string(s)
}

var registry Registry

type Registry struct {
	collectors sync.Map
}

func (c *Registry) Register(tpy string, builder Builder) {
	c.collectors.Store(tpy, builder)
}

func (c *Registry) GetCollector(ctx context.Context, rawConf json.RawMessage) Collector {
	type collectorTypeConf struct {
		Type string `json:"type"`
	}

	typeConf := &collectorTypeConf{}
	if err := json.Unmarshal(rawConf, typeConf); err != nil {
		log.Panic("load config fail", err)
	}

	builder, exist := c.collectors.Load(typeConf.Type)
	if !exist {
		return nil
	}

	return builder.(Builder)(ctx, rawConf)
}

func GetRegistry() *Registry {
	return &registry
}
