package config

import fwconfig "github.com/zhoudm1743/go-fast-framework/config"

func init() {
	fwconfig.Add("queue", map[string]any{
		// 同步驱动 worker 池大小（0 表示使用 runtime.NumCPU()）
		"workers": 0,
	})
}
