# Dify Gateway

这是一个由JiuXia2025编写的基于 Gin 框架的 Dify API 网关服务，用于将 Dify 格式的 API 请求转换为 OpenAI API 请求，支持模型交互时流式传输。

## 功能特点

- 支持 OpenAI 格式的 API 请求
- 支持流式响应
- 支持模型列表查询
- 支持跨域请求
- 可配置的 API 地址和密钥
- 可配置的模型信息

## 配置说明

配置文件 `app.yaml` 包含以下配置项：

```yaml
dify:
  api_key: app-xxxxxxxxxxxxxxxxxx  # Dify API 密钥
  api_url: https://api.dify.ai/v1  # Dify API 地址

model:
  id: Deepseek-Dify      # 模型 ID
  object: model          # 模型类型（不建议修改）
  created: 1686935002    # 创建时间
  owned_by: JiuXia2025   # 模型所有者

server:
  port: 8089            # 服务端口
```

## 安装和使用

#### 方式一：自己构建

1. 克隆项目：

```bash
git clone https://github.com/JiuXia2025/dify-gateway.git
cd dify-gateway
```

2. 安装依赖：

```bash
go mod download
```

3. 修改配置文件：

编辑 `app.yaml` 文件，填入您的 Dify API 密钥和其他配置。

4. 运行服务：

```bash
go run main.go
```

服务将在配置的端口（默认 8089）启动。

#### 方法二：使用发布的构建成品(无需构建)

1. 下载软件：

前往[Releases](https://github.com/JiuXia2025/dify-gateway/releases)页面下载对应系统的构建成品

2. 修改配置文件：

编辑 `app.yaml` 文件，填入您的 Dify API 密钥和其他配置。

3. 运行服务：

```bash
Linux：
./dify-gateway

Windows：
.\dify-gateway.exe
或双击启动exe
```

服务将在配置的端口（默认 8089）启动。

## API 接口

### 获取模型列表

```
GET /v1/models
```

响应示例：

```json
{
  "object": "list",
  "data": [
    {
      "id": "Deepseek-Dify",
      "object": "model",
      "created": 1686935002,
      "owned_by": "JiuXia2025"
    }
  ]
}
```

### 聊天补全

```
POST /v1/chat/completions
```

请求示例：

```json
{
  "model": "Deepseek-Dify",
  "messages": [
    {
      "role": "user",
      "content": "你好"
    }
  ],
  "stream": true
}
```

## 注意事项

1. 请确保您的 Dify API 密钥有效
2. 如果需要修改模型信息，请更新配置文件中的相应字段
3. 服务默认监听 8089 端口，可以在配置文件中修改

## 许可证

MIT License
