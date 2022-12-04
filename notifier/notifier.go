package notifier

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"sync"
)

type Notifier interface {
	Notify(msg, from string, throttle bool)
}

type Builder func(ctx context.Context, rawConf json.RawMessage) Notifier

var registry Registry

type Registry struct {
	notifiers sync.Map
}

func (r *Registry) Register(tpy string, builder Builder) {
	r.notifiers.Store(tpy, builder)
}

func (r *Registry) GetNotifier(ctx context.Context, rawConf json.RawMessage) Notifier {
	type collectorTypeConf struct {
		Type string `json:"type"`
	}

	typeConf := &collectorTypeConf{}
	if err := json.Unmarshal(rawConf, typeConf); err != nil {
		log.Panic("load config fail", err)
	}

	builder, exist := r.notifiers.Load(typeConf.Type)
	if !exist {
		return nil
	}

	return builder.(Builder)(ctx, rawConf)
}

func GetRegistry() *Registry {
	return &registry
}
