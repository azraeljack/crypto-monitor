package strategy

import (
	"context"
	"encoding/json"
	"github.com/azraeljack/crypto-monitor/collector"
	"github.com/azraeljack/crypto-monitor/notifier"
	log "github.com/sirupsen/logrus"
	"sync"
)

type Builder func(ctx context.Context, rawConf json.RawMessage) Strategy

type Strategy interface {
	Run()
	AddCollectors(collector ...collector.Collector)
	AddNotifiers(notifier ...notifier.Notifier)
}

var registry Registry

type Registry struct {
	strategies sync.Map
}

func (r *Registry) Register(tpy string, builder Builder) {
	r.strategies.Store(tpy, builder)
}

func (r *Registry) GetStrategy(ctx context.Context, rawConf json.RawMessage) Strategy {
	type collectorTypeConf struct {
		Type string `json:"type"`
	}

	typeConf := &collectorTypeConf{}
	if err := json.Unmarshal(rawConf, typeConf); err != nil {
		log.Panic("load config fail", err)
	}

	builder, exist := r.strategies.Load(typeConf.Type)
	if !exist {
		return nil
	}

	return builder.(Builder)(ctx, rawConf)
}

func GetRegistry() *Registry {
	return &registry
}
