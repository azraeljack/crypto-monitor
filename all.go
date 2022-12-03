package main

import (
	// collectors
	_ "github.com/azraeljack/crypto-monitor/collector/binance"

	// notifiers
	_ "github.com/azraeljack/crypto-monitor/notifier/wechat"

	// strategies
	_ "github.com/azraeljack/crypto-monitor/strategy/price_change"
)