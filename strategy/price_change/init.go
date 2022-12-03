package price_change

import (
	"github.com/azraeljack/crypto-monitor/strategy"
)

func init() {
	strategy.GetRegistry().Register("price_change", NewPriceChangeStrategy)
}
