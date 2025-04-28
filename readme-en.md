# Dify Gateway

[简体中文](./readme.md) | [English](./readme-en.md)

The project is a gateway service which converts Dify API requests to OpenAI API requests, supports model interaction with streaming based on Gin framework. Developed by JiuXia2025.

## Features

- Support OpenAI API requests
- Support model interaction with streaming
- Support model list query
- Support cross-origin requests
- Configurable API address and key
- Configurable model information

## Configuration

The `app.yaml` includes the following configuration items

```yaml
dify:
  api_key: app-xxxxxxxxxxxxxxxxxx # Your Dify API KEY
  api_url: https://api.dify.ai/v1 # Your Dify API URL

model:
  id: Deepseek-Dify # Model ID
  object: model # Model type (No changes recommended)
  created: 1686935002 # Creation Time (unix timestamp)
  owned_by: JiuXia2025 # Model owner

server:
  port: 8089 # Service port
```

## Installation

#### I. Build from Source

1. Clone the repository

```bash
git clone https://github.com/JiuXia2025/dify-gateway.git
cd dify-gateway
```

2. Install dependencies

```bash
go mod download
```

3. Modify the configuration

Edit `app.yaml` file and fill in your Dify API KEY and other configurations.

4. Run the service

```bash
go run main.go
```

Service will start on the configured port (default 8089).

#### II. Download and run app directly from release (no build required)

1. Download app from the release

Go to [Releases](https://github.com/JiuXia2025/dify-gateway/releases) page to download the release for your system.

2. Modify the configuration

Edit `app.yaml` file and fill in your Dify API KEY and other configurations.

3. Run the service

```bash
Linux：
./dify-gateway

Windows：
.\dify-gateway.exe
or double click to run exe
```

Service will start on the configured port (default 8089).

## API Interfaces

### Get Model List

```
GET /v1/models
```

Response Example:

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

### Chat Completion

```
POST /v1/chat/completions
```

Request Example:

```json
{
  "model": "Deepseek-Dify",
  "messages": [
    {
      "role": "user",
      "content": "Hello"
    }
  ],
  "stream": true
}
```

## The important information you should know

1. Make sure your Dify API key is valid
2. If you need to modify model information, please update the corresponding fields in the configuration file
3. The service listens on port 8089 by default, which can be modified in the configuration file

## License

MIT License
