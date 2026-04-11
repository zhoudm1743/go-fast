# 文件存储

## 简介

GoFast 文件存储模块提供了统一的文件操作接口，支持本地磁盘、阿里云 OSS、腾讯云 COS、MinIO 和 AWS S3 等多种驱动。  
使用 `facades.Storage()` 操作文件存储。

---

## 配置

在 `config/config.yaml` 中配置文件系统：

```yaml
filesystem:
  default: local       # 默认磁盘
  disks:
    local:
      driver: local
      root: storage/app   # 本地存储根目录
      url: /storage       # 文件访问 URL 前缀

    oss:
      driver: oss
      key: your-access-key
      secret: your-access-secret
      bucket: your-bucket-name
      endpoint: oss-cn-hangzhou.aliyuncs.com
      url: https://your-bucket.oss-cn-hangzhou.aliyuncs.com

    cos:
      driver: cos
      key: your-secret-id
      secret: your-secret-key
      url: https://your-bucket-1234567890.cos.ap-guangzhou.myqcloud.com

    minio:
      driver: minio
      key: your-access-key
      secret: your-secret-key
      bucket: your-bucket
      endpoint: http://127.0.0.1:9000
      url: http://127.0.0.1:9000/your-bucket
      region: us-east-1
      ssl: false

    s3:
      driver: s3
      key: your-access-key
      secret: your-secret-key
      region: us-east-1
      bucket: your-bucket
      url: https://your-bucket.s3.us-east-1.amazonaws.com
      endpoint: ""              # 自定义端点（可选）
      cdn: ""                   # CDN 域名（可选）
      token: ""                 # Session Token（可选）
      object_canned_acl: ""     # 如 public-read（可选）
      use_path_style: false
```

---

## 注册服务提供者

在 `bootstrap/app.go` 中注册 Filesystem 服务提供者：

```go
import "github.com/zhoudm1743/go-fast/framework/filesystem"

app.Register(&filesystem.ServiceProvider{})
```

---

## 基本用法

通过 `facades.Storage()` 获取默认磁盘实例：

```go
import "github.com/zhoudm1743/go-fast/framework/facades"

storage := facades.Storage()
```

切换到指定磁盘：

```go
storage := facades.Storage().Disk("oss")
```

---

## 文件操作

### 写入文件

```go
// 写入字符串内容
err := facades.Storage().Put("path/to/file.txt", "Hello, GoFast!")

// 上传文件对象（随机生成文件名）
key, err := facades.Storage().PutFile("uploads", file)

// 上传文件对象并指定文件名
key, err := facades.Storage().PutFileAs("uploads", file, "avatar.png")
```

### 读取文件

```go
// 读取为字符串
content, err := facades.Storage().Get("path/to/file.txt")

// 读取为字节数组
data, err := facades.Storage().GetBytes("path/to/file.txt")
```

### 判断文件是否存在

```go
exists := facades.Storage().Exists("path/to/file.txt")
missing := facades.Storage().Missing("path/to/file.txt")
```

### 获取文件 URL

```go
// 获取公开访问 URL
url := facades.Storage().Url("path/to/file.txt")

// 获取临时 URL（过期时间戳，Unix 秒）
tempUrl, err := facades.Storage().TemporaryUrl("path/to/file.txt", time.Now().Add(time.Hour).Unix())
```

### 复制与移动

```go
// 复制文件
err := facades.Storage().Copy("old/path.txt", "new/path.txt")

// 移动文件
err := facades.Storage().Move("old/path.txt", "new/path.txt")
```

### 删除文件

```go
// 删除单个文件
err := facades.Storage().Delete("path/to/file.txt")

// 批量删除文件
err := facades.Storage().Delete("file1.txt", "file2.txt", "file3.txt")
```

### 文件元信息

```go
// 获取文件大小（字节）
size, err := facades.Storage().Size("path/to/file.txt")

// 获取最后修改时间（Unix 时间戳）
ts, err := facades.Storage().LastModified("path/to/file.txt")

// 获取 MIME 类型
mime, err := facades.Storage().MimeType("path/to/file.txt")

// 获取文件完整路径（仅本地驱动有意义）
fullPath := facades.Storage().Path("path/to/file.txt")
```

---

## 目录操作

```go
// 创建目录
err := facades.Storage().MakeDirectory("uploads/images")

// 删除目录（递归）
err := facades.Storage().DeleteDirectory("uploads/temp")

// 列出目录下的文件（非递归）
files, err := facades.Storage().Files("uploads")

// 递归列出目录下所有文件
files, err := facades.Storage().AllFiles("uploads")

// 列出目录下的子目录（非递归）
dirs, err := facades.Storage().Directories("uploads")

// 递归列出所有子目录
dirs, err := facades.Storage().AllDirectories("uploads")
```

---

## 上下文支持

所有驱动均支持传递 `context.Context`，用于超时控制和请求追踪：

```go
import "context"

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := facades.Storage().WithContext(ctx).Put("path/to/file.txt", "content")
```

---

## 驱动说明

| 驱动    | 配置 `driver` 值 | 必填字段                                         |
|---------|-----------------|--------------------------------------------------|
| 本地    | `local`         | `root`, `url`                                    |
| 阿里云 OSS | `oss`       | `key`, `secret`, `bucket`, `endpoint`, `url`     |
| 腾讯云 COS | `cos`       | `key`, `secret`, `url`                           |
| MinIO   | `minio`         | `key`, `secret`, `bucket`, `endpoint`, `url`     |
| AWS S3  | `s3`            | `key`, `secret`, `region`, `bucket`, `url`       |

> 本地驱动的 `TemporaryUrl` 直接返回普通 URL，不做过期控制。
