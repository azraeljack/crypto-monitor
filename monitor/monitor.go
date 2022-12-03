package monitor

import (
	"context"
	"encoding/json"
	"github.com/azraeljack/crypto-monitor/collector"
	"github.com/azraeljack/crypto-monitor/config"
	"github.com/azraeljack/crypto-monitor/notifier"
	"github.com/azraeljack/crypto-monitor/strategy"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"strings"
)

type Monitor struct {
	collectors []collector.Collector
	notifiers  []notifier.Notifier
	strategies []strategy.Strategy

	ctx context.Context
}

func NewMonitor(ctx context.Context, configFile string) *Monitor {
	if strings.HasPrefix(configFile, ".") {
		cwd, _ := os.Getwd()
		configFile = path.Join(cwd, configFile)
	}

	log.Infof("loading config file: %s ...", configFile)
	f, err := os.Open(configFile)
	if err != nil {
		log.Panicf("failed to load config: %v", err)
	}

	rawConf, err := io.ReadAll(f)
	if err != nil {
		log.Panicf("failed to read config: %v", err)
	}

	conf := &config.Config{}
	if err := json.Unmarshal(rawConf, conf); err != nil {
		log.Panicf("failed to parse config: %v", err)
	}

	monitor := &Monitor{
		ctx: ctx,
	}

	for _, collectorConf := range conf.Collectors {
		monitor.collectors = append(monitor.collectors, collector.GetRegistry().GetCollector(ctx, collectorConf))
	}

	for _, notifierConf := range conf.Notifiers {
		monitor.notifiers = append(monitor.notifiers, notifier.GetRegistry().GetNotifier(ctx, notifierConf))
	}

	for _, strategyConf := range conf.Strategies {
		strata := strategy.GetRegistry().GetStrategy(ctx, strategyConf)
		strata.AddNotifiers(monitor.notifiers...)
		strata.AddCollectors(monitor.collectors...)

		monitor.strategies = append(monitor.strategies, strata)
	}

	return monitor
}

func (m *Monitor) Start() {
	for _, strata := range m.strategies {
		strata.Run()
	}
}
