package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// 配置文件结构
type AppConfig struct {
	Dify struct {
		APIKey string `yaml:"api_key"`
		APIURL string `yaml:"api_url"`
	} `yaml:"dify"`
	Model struct {
		ID      string `yaml:"id"`
		Object  string `yaml:"object"`
		Created int64  `yaml:"created"`
		OwnedBy string `yaml:"owned_by"`
	} `yaml:"model"`
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
}

var config AppConfig

// OpenAI 请求响应结构体
type OpenAIRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	User     string          `json:"user"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIStreamResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   interface{}    `json:"usage,omitempty"`
}

type OpenAIChoice struct {
	Index        int         `json:"index"`
	Delta        OpenAIDelta `json:"delta"`
	FinishReason interface{} `json:"finish_reason,omitempty"`
}

type OpenAIDelta struct {
	Content string `json:"content,omitempty"`
}

// Dify 请求响应结构体
type DifyRequest struct {
	Query          string                 `json:"query"`
	Inputs         map[string]interface{} `json:"inputs"`
	ResponseMode   string                 `json:"response_mode"`
	User           string                 `json:"user"`
	ConversationID string                 `json:"conversation_id,omitempty"`
}

type DifyEvent struct {
	Event          string                 `json:"event"`
	TaskID         string                 `json:"task_id,omitempty"`
	MessageID      string                 `json:"message_id,omitempty"`
	Answer         string                 `json:"answer,omitempty"`
	CreatedAt      int64                  `json:"created_at,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	ConversationID string                 `json:"conversation_id,omitempty"`
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Type")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
func main() {
	// 加载配置文件
	if err := loadConfig("app.yaml"); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化Gin
	r := gin.Default()
	r.Use(CORSMiddleware())
	r.Use(loggerMiddleware())

	// 注册路由
	r.GET("/v1/models", handleModels)
	r.POST("/v1/chat/completions", handleChatCompletion)

	// 添加静态文件服务
	r.Static("/app", "./dist")

	// 启动服务
	port := fmt.Sprintf(":%d", config.Server.Port)
	r.Run(port)
}

//func main() {
//	// 加载配置文件
//	if err := loadConfig("app.yaml"); err != nil {
//		fmt.Printf("Failed to load config: %v\n", err)
//		os.Exit(1)
//	}
//
//	// 初始化Gin
//	r := gin.Default()
//	r.Use(CORSMiddleware())
//	r.Use(loggerMiddleware())
//
//	// 注册路由
//	r.GET("/v1/models", handleModels)
//	r.POST("/v1/chat/completions", handleChatCompletion)
//
//	//r.Static("/app", "./dist")
//
//	// 启动服务
//	port := fmt.Sprintf(":%d", config.Server.Port)
//	r.Run(port)
//}

func loadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("open config file: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	// 验证必要配置
	if config.Dify.APIKey == "" {
		return fmt.Errorf("missing dify.api_key in config")
	}
	if config.Dify.APIURL == "" {
		return fmt.Errorf("missing dify.api_url in config")
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}

	return nil
}

func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		fmt.Printf("[%s] %s %s - %v - %d\n",
			time.Now().Format("2006-01-02 15:04:05"),
			c.Request.Method,
			c.Request.URL.Path,
			duration,
			c.Writer.Status(),
		)
	}
}

func handleModels(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data": []gin.H{
			{
				"id":       config.Model.ID,
				"object":   config.Model.Object,
				"created":  config.Model.Created,
				"owned_by": config.Model.OwnedBy,
			},
		},
	})
}

func handleChatCompletion(c *gin.Context) {
	var openAIReq OpenAIRequest
	if err := c.ShouldBindJSON(&openAIReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	difyReq, err := convertToDifyRequest(openAIReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建Dify请求
	reqBody, _ := json.Marshal(difyReq)
	req, _ := http.NewRequest("POST", config.Dify.APIURL+"/chat-messages", bytes.NewBuffer(reqBody))
	req.Header.Set("Authorization", "Bearer "+config.Dify.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to connect to Dify API"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.JSON(resp.StatusCode, gin.H{"error": string(body)})
		return
	}

	if openAIReq.Stream {
		handleStreamResponse(c, resp.Body, openAIReq)
	} else {
		handleBlockingResponse(c, resp.Body, openAIReq)
	}
}

func convertToDifyRequest(req OpenAIRequest) (*DifyRequest, error) {
	var query string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			query = req.Messages[i].Content
			break
		}
	}
	if query == "" {
		return nil, fmt.Errorf("no user message found")
	}

	mode := "blocking"
	if req.Stream {
		mode = "streaming"
	}

	return &DifyRequest{
		Query:        query,
		Inputs:       make(map[string]interface{}),
		ResponseMode: mode,
		User:         req.User,
	}, nil
}

func handleStreamResponse(c *gin.Context, body io.ReadCloser, req OpenAIRequest) {
	// 生成唯一请求ID（示例使用时间戳+用户）
	requestID := fmt.Sprintf("%d-%s", time.Now().UnixNano(), req.User)
	startTime := time.Now()

	// 记录请求开始
	log.Printf("[%s] Stream request started | Model: %s | Stream: %t",
		requestID, req.Model, req.Stream)

	defer func() {
		// 记录请求结束
		duration := time.Since(startTime)
		log.Printf("[%s] Stream request completed | Duration: %s | Status: %s",
			requestID, duration.Round(time.Millisecond), getStatus(c))
	}()

	c.Stream(func(w io.Writer) bool {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")

		reader := bufio.NewReader(body)
		var (
			fullContent  string
			eventCounter int
		)

		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					// 正常结束日志
					log.Printf("[%s] Received EOF | Total events: %d | Content length: %d",
						requestID, eventCounter, len(fullContent))
					sendStreamEnd(w, req, fullContent)
					return false
				}
				// 错误日志
				log.Printf("[%s] Read error: %v | Bytes read: %d",
					requestID, err, len(line))
				return false
			}

			rawLine := bytes.TrimSpace(line)
			if len(rawLine) == 0 {
				continue
			}

			// 记录原始数据（调试用，生产环境建议使用DEBUG级别）
			log.Printf("[%s] Raw data received: %s", requestID, string(rawLine))

			if !bytes.HasPrefix(rawLine, []byte("data: ")) {
				log.Printf("[%s] Unexpected data format: %s",
					requestID, string(rawLine))
				continue
			}

			var event DifyEvent
			jsonData := bytes.TrimPrefix(rawLine, []byte("data: "))

			if err := json.Unmarshal(jsonData, &event); err != nil {
				// JSON解析错误日志
				log.Printf("[%s] JSON parse failed | Error: %v | Data: %s",
					requestID, err, string(jsonData))
				continue
			}

			// 记录事件基本信息
			log.Printf("[%s] Event received | Type: %s | MessageID: %s",
				requestID, event.Event, event.MessageID)

			eventCounter++

			switch event.Event {
			case "message":
				fullContent += event.Answer
				log.Printf("[%s] Message chunk | Length: %d | Content: %.40s...",
					requestID, len(event.Answer), event.Answer)

				if err := sendStreamChunk(w, event, req.Model); err != nil {
					log.Printf("[%s] Send chunk failed: %v", requestID, err)
					return false
				}

			case "message_end":
				// 最终日志
				log.Printf("[%s] Message end received | Total length: %d | Metadata: %+v",
					requestID, len(fullContent), event.Metadata)
				sendStreamEnd(w, req, fullContent)
				return false

			case "error":
				log.Printf("[%s] Error event | Code: %v | Message: %v",
					requestID, event.Metadata["code"], event.Metadata["message"])
				sendErrorChunk(w, event)
				return false

			default:
				log.Printf("[%s] Unknown event type: %s", requestID, event.Event)
			}
		}
	})
}

// Helper function to get request status
func getStatus(c *gin.Context) string {
	if c.Writer.Written() {
		return "completed"
	}
	if c.IsAborted() {
		return "aborted"
	}
	return "unknown"
}

// 修改send方法返回error
func sendStreamChunk(w io.Writer, event DifyEvent, model string) error {
	resp := OpenAIStreamResponse{
		ID:      event.MessageID,
		Object:  "chat.completion.chunk",
		Created: event.CreatedAt,
		Model:   model,
		Choices: []OpenAIChoice{{
			Index: 0,
			Delta: OpenAIDelta{Content: event.Answer},
		}},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}

func sendStreamEnd(w io.Writer, req OpenAIRequest, content string) {
	resp := OpenAIStreamResponse{
		ID:      fmt.Sprintf("chatcmpl-%d", time.Now().Unix()),
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []OpenAIChoice{{
			Index:        0,
			FinishReason: "stop",
		}},
	}

	data, _ := json.Marshal(resp)
	fmt.Fprintf(w, "data: %s\n\n", data)
	fmt.Fprintf(w, "data: [DONE]\n\n")
	w.(http.Flusher).Flush()
}

func sendErrorChunk(w io.Writer, event DifyEvent) {
	errData := map[string]interface{}{
		"error": map[string]interface{}{
			"message": "Dify API error",
			"code":    event.Metadata["status"],
			"type":    "api_error",
		},
	}
	data, _ := json.Marshal(errData)
	fmt.Fprintf(w, "data: %s\n\n", data)
	w.(http.Flusher).Flush()
}

func handleBlockingResponse(c *gin.Context, body io.ReadCloser, req OpenAIRequest) {
	var difyResp DifyEvent
	if err := json.NewDecoder(body).Decode(&difyResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response"})
		return
	}
	//流式缓冲处理开始
	c.Stream(func(w io.Writer) bool {
		// 禁用框架层面的缓冲
		c.Writer.Flush()

		reader := bufio.NewReader(body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				return false
			}

			// 直接转发原始数据块
			fmt.Fprintf(w, "%s", line)
			c.Writer.Flush() // 立即刷新缓冲区
		}
	})
	//流式缓冲处理结束
	c.JSON(http.StatusOK, gin.H{
		"id":      difyResp.MessageID,
		"object":  "chat.completion",
		"created": difyResp.CreatedAt,
		"model":   req.Model,
		"choices": []gin.H{{
			"index": 0,
			"message": gin.H{
				"role":    "assistant",
				"content": difyResp.Answer,
			},
		}},
		"usage": difyResp.Metadata["usage"],
	})
}

func logError(context string, err error) {
	fmt.Printf("[ERROR] %s: %v\n", context, err)
}
