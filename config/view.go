package config

import fwconfig "github.com/zhoudm1743/go-fast-framework/config"

func init() {
	fwconfig.Add("view", map[string]any{
		// 模板根目录（相对于项目根）。留空则禁用视图引擎。
		"dir": "resources/views",
		// 文件扩展名过滤，空字符串表示加载全部
		"extension": ".html",
		// 开发模式热重载（生产环境改为 false）
		"reload": true,
	})
}
