package binance

import "github.com/azraeljack/crypto-monitor/collector"

func init() {
	collector.GetRegistry().Register("binance", NewBinanceCollector)
}
