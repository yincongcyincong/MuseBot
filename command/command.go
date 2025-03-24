package command

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/cohesion-org/deepseek-go/constants"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/yincongcyincong/telegram-deepseek-bot/conf"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

// CostumCommand command struct
type CostumCommand struct {
	Crontab   string                 `json:"crontab"`
	Command   string                 `json:"command"`
	SendUser  string                 `json:"send_user"`
	SendGroup string                 `json:"send_group"`
	Chains    []*Chain               `json:"chains"` // chain list
	Param     map[string]interface{} `json:"param"`
}

// Chain chain struct
type Chain struct {
	Type  string  `json:"type"`  // 任务类型（http 或 deepseek）
	Tasks []*Task `json:"tasks"` // 任务列表
	Proxy string  `json:"proxy"`
}

// Task task
type Task struct {
	Name      string     `json:"name"`                 // 任务名称
	HTTPParam *HTTPParam `json:"http_param,omitempty"` // HTTP 请求参数（仅当 type 为 http 时有效）
	Template  string     `json:"template,omitempty"`   // 模板字符串（仅当 type 为 deepseek 时有效）
}

const (
	TaskTypeHTTP     = "http"
	TaskTypeDeepseek = "deepseek"
)

// HTTPParam send http request
type HTTPParam struct {
	URL     string            `json:"url"`     // 请求 URL
	Method  string            `json:"method"`  // 请求方法（GET、POST 等）
	Headers map[string]string `json:"headers"` // 请求头
	Body    string            `json:"body"`    // 请求体
}

var (
	CustomeCommandList = make([]*CostumCommand, 0)
)

func LoadCustomCommands() {
	file, err := os.Open("./command/command.json")
	if err != nil {
		logger.Error("open command.json error", err)
		return
	}
	defer file.Close()

	// 读取文件内容
	data, err := ioutil.ReadAll(file)
	if err != nil {
		logger.Error("read command.json error", err)
		return
	}

	// 将 JSON 解析到结构体
	err = json.Unmarshal(data, &CustomeCommandList)
	if err != nil {
		logger.Error("parse command.json error", err)
	}
}

func ExecuteCustomCommand(command string) {
	var c *CostumCommand
	for _, customCommand := range CustomeCommandList {
		if customCommand.Command == command {
			c = customCommand
			break
		}
	}

	if c == nil {
		return
	}

	for _, chain := range c.Chains {
		switch chain.Type {
		case TaskTypeHTTP:
			c.concurrentHTTPRequests(chain.Tasks, chain.Proxy)
		case TaskTypeDeepseek:
			c.sendDeepseekContent(chain.Tasks, chain.Proxy)
		}
	}
}

// fetchURL send http request and send result to channel
func (c *CostumCommand) fetchURL(config *HTTPParam, wg *sync.WaitGroup, resultChan chan<- map[string]interface{}, key, proxyUrl string) {
	defer wg.Done()

	t, err := template.New("deepseek_content").Parse(config.URL)
	if err != nil {
		logger.Warn("parse template error", "err", err)
		return
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, c.Param)
	if err != nil {
		logger.Warn("render template error", "err", err)
		return
	}

	urlStr := buf.String()

	req, err := http.NewRequest(config.Method, urlStr, bytes.NewBuffer([]byte(config.Body)))
	if err != nil {
		resultChan <- map[string]interface{}{key: fmt.Sprintf("Error: %v", err)}
		return
	}

	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	if proxyUrl != "" {
		proxy, err := url.Parse(proxyUrl)
		if err != nil {
			logger.Error("parse http request proxy error", "err", err)
		} else {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		resultChan <- map[string]interface{}{key: fmt.Sprintf("Error: %v", err)}
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resultChan <- map[string]interface{}{key: fmt.Sprintf("Error: %v", err)}
		return
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		resultChan <- map[string]interface{}{key: fmt.Sprintf("Error: %v", err)}
		return
	}

	resultChan <- map[string]interface{}{key: data}
}

// concurrentHTTPRequests send http requests concurrently
func (c *CostumCommand) concurrentHTTPRequests(requests []*Task, proxy string) {
	var wg sync.WaitGroup
	resultChan := make(chan map[string]interface{}, len(requests))

	for _, config := range requests {
		wg.Add(1)
		go c.fetchURL(config.HTTPParam, &wg, resultChan, config.Name, proxy)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	for resp := range resultChan {
		for key, value := range resp {
			c.Param[key] = value
		}
	}

}

func (c *CostumCommand) sendDeepseekContent(requests []*Task, proxy string) {
	for _, config := range requests {
		question, answer, err := c.getDeepseekContent(config.Template, proxy)
		if err != nil {
			logger.Error("getDeepseekContent error", "err", err)
			continue
		}

		userIDs := strings.Split(c.SendUser, ",")
		for _, userID := range userIDs {
			userIDInt, err := strconv.ParseInt(userID, 10, 64)
			if err != nil {
				logger.Error("parse userID error", "err", err)
				continue
			}
			sendMsg(userIDInt, config.Name+" question: "+question)
			sendMsg(userIDInt, config.Name+" answer: "+answer)
		}
	}
}

func (c *CostumCommand) getDeepseekContent(tpl string, proxyUrl string) (string, string, error) {
	t, err := template.New("deepseek_content").Parse(tpl)
	if err != nil {
		logger.Warn("parse template error", "err", err)
		return "", "", err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, c.Param)
	if err != nil {
		logger.Warn("render template error", "err", err)
		return "", "", err
	}

	messages := make([]deepseek.ChatCompletionMessage, 0)
	messages = append(messages, deepseek.ChatCompletionMessage{
		Role:    constants.ChatMessageRoleUser,
		Content: buf.String(),
	})

	request := &deepseek.ChatCompletionRequest{
		Model:    deepseek.DeepSeekChat,
		Messages: messages,
	}
	ctx := context.Background()
	httpClient := &http.Client{
		Timeout: 30 * time.Minute,
	}

	if proxyUrl != "" {
		proxy, err := url.Parse(proxyUrl)
		if err != nil {
			logger.Error("parse deepseek proxy error", "err", err)
		} else {
			httpClient.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}

	client, err := deepseek.NewClientWithOptions(*conf.DeepseekToken,
		deepseek.WithBaseURL(*conf.CustomUrl), deepseek.WithHTTPClient(httpClient))

	logger.Info("msg create", "prompt", buf.String())
	response, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		logger.Error("ChatCompletionStream error", "err", err)
		return "", "", err
	}

	content := ""
	for _, choice := range response.Choices {
		content += choice.Message.Content
	}

	return buf.String(), content, nil
}

func sendMsg(userId int64, content string) {
	tgMsgInfo := tgbotapi.NewMessage(userId, content)
	tgMsgInfo.ParseMode = tgbotapi.ModeMarkdown
	_, err := conf.Bot.Send(tgMsgInfo)
	if err != nil {
		logger.Warn("Sending first message fail", "err", err)
	}
}
