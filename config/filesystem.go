package config

import fwconfig "github.com/zhoudm1743/go-fast-framework/config"

func init() {
	fwconfig.Add("filesystem", map[string]any{
		// 默认磁盘名称
		"default": "local",

		// 磁盘配置
		"disks": map[string]any{
			// 本地存储
			"local": map[string]any{
				"driver": "local",
				"root":   "storage/app",
				"url":    "/storage",
			},

			// 阿里云 OSS
			// "oss": map[string]any{
			//     "driver":   "oss",
			//     "key":      "your-access-key",
			//     "secret":   "your-access-secret",
			//     "bucket":   "your-bucket",
			//     "url":      "https://your-bucket.oss-cn-hangzhou.aliyuncs.com",
			//     "endpoint": "oss-cn-hangzhou.aliyuncs.com",
			// },

			// 腾讯云 COS
			// "cos": map[string]any{
			//     "driver": "cos",
			//     "key":    "your-secret-id",
			//     "secret": "your-secret-key",
			//     "url":    "https://your-bucket.cos.ap-guangzhou.myqcloud.com",
			// },

			// MinIO
			// "minio": map[string]any{
			//     "driver":   "minio",
			//     "key":      "your-access-key",
			//     "secret":   "your-secret-key",
			//     "bucket":   "your-bucket",
			//     "url":      "http://localhost:9000",
			//     "endpoint": "localhost:9000",
			//     "region":   "us-east-1",
			//     "ssl":      false,
			// },

			// AWS S3
			// "s3": map[string]any{
			//     "driver":           "s3",
			//     "key":              "your-access-key",
			//     "secret":           "your-secret-key",
			//     "region":           "us-east-1",
			//     "bucket":           "your-bucket",
			//     "url":              "https://your-bucket.s3.amazonaws.com",
			//     "endpoint":         "",
			//     "cdn":              "",
			//     "use_path_style":   false,
			//     "object_canned_acl": "private",
			// },
		},
	})
}
