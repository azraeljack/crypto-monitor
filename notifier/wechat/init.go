package wechat

import (
	"github.com/azraeljack/crypto-monitor/notifier"
)

func init() {
	notifier.GetRegistry().Register("wechat", NewWechatNotifier)
}
