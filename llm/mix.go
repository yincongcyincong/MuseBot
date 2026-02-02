package llm

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	
	"github.com/cohesion-org/deepseek-go/constants"
	openrouter "github.com/revrost/go-openrouter"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

var (
	pngFetch    = regexp.MustCompile(`https?://[^\s)]+\.png[^\s)]*`)
	mkBase64Reg = regexp.MustCompile(`!\[.*?\]\((data:image\/[a-zA-Z0-9\+\-\.]*;base64,([a-zA-Z0-9\+\/\=]+))\)`)
)

// AI302FetchResp fetch response
type AI302FetchResp struct {
	TaskID         string           `json:"task_id"`
	UpstreamTaskID string           `json:"upstream_task_id"`
	Status         string           `json:"status"`
	VideoURL       string           `json:"video_url"`
	RawResponse    AI302RawResponse `json:"raw_response"`
	Model          string           `json:"model"`
	ExecutionTime  int              `json:"execution_time"`
	CreatedAt      *string          `json:"created_at"`
	CompletedAt    string           `json:"completed_at"`
}

type AI302RawResponse struct {
	CreatedAt string `json:"created_at"`
	ID        string `json:"id"`
	Prompt    string `json:"prompt"`
	State     string `json:"state"`
	Video     string `json:"video"`
}

// Create video response
type CreateResp struct {
	TaskID string `json:"task_id"`
}

func GenerateMixImg(ctx context.Context, prompt string, imageContent []byte) ([]byte, int, error) {
	start := time.Now()
	llmConfig := db.GetCtxUserInfo(ctx).LLMConfigRaw
	mediaType := utils.GetImgType(llmConfig)
	model := utils.GetUsingImgModel(mediaType, llmConfig.ImgModel)
	metrics.APIRequestCount.WithLabelValues(model).Inc()
	
	messages := openrouter.ChatCompletionMessage{
		Role: constants.ChatMessageRoleUser,
		Content: openrouter.Content{
			Multi: []openrouter.ChatMessagePart{
				{
					Type: openrouter.ChatMessagePartTypeText,
					Text: prompt,
				},
			},
		},
	}
	
	if len(imageContent) != 0 {
		messages.Content.Multi = append(messages.Content.Multi, openrouter.ChatMessagePart{
			Type: openrouter.ChatMessagePartTypeImageURL,
			ImageURL: &openrouter.ChatMessageImageURL{
				URL: "data:image/" + utils.DetectImageFormat(imageContent) + ";base64," + base64.StdEncoding.EncodeToString(imageContent),
			},
		})
	}
	
	client := GetMixClient(ctx, "img")
	request := openrouter.ChatCompletionRequest{
		Model:    model,
		Messages: []openrouter.ChatCompletionMessage{messages},
	}
	
	// assign task
	var response openrouter.ChatCompletionResponse
	var err error
	for i := 0; i < conf.BaseConfInfo.LLMRetryTimes; i++ {
		response, err = client.CreateChatCompletion(ctx, request)
		if err != nil {
			time.Sleep(time.Duration(conf.BaseConfInfo.LLMRetryInterval) * time.Millisecond)
			continue
		}
		break
	}
	
	if err != nil {
		logger.ErrorCtx(ctx, "create chat completion fail", "err", err)
		return nil, 0, err
	}
	
	metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())
	
	if len(response.Choices) != 0 {
		if len(response.Choices[0].Message.Content.Multi) != 0 {
			imageContent, err := utils.DownloadFile(response.Choices[0].Message.Content.Multi[0].ImageURL.URL)
			if err != nil {
				logger.ErrorCtx(ctx, "download image fail", "err", err)
				return nil, 0, err
			}
			return imageContent, response.Usage.TotalTokens, nil
		} else if strings.Contains(response.Choices[0].Message.Content.Text, "http") {
			pngs := pngFetch.FindAllString(response.Choices[0].Message.Content.Text, -1)
			imageContent, err := utils.DownloadFile(pngs[len(pngs)-1])
			if err != nil {
				logger.ErrorCtx(ctx, "download image fail", "err", err)
				return nil, 0, err
			}
			return imageContent, response.Usage.TotalTokens, nil
		} else if strings.Contains(response.Choices[0].Message.Content.Text, "data:image") {
			matches := mkBase64Reg.FindAllStringSubmatch(response.Choices[0].Message.Content.Text, -1)
			if len(matches) > 0 && len(matches[0]) > 2 {
				b64, err := base64.StdEncoding.DecodeString(matches[0][2])
				if err != nil {
					logger.ErrorCtx(ctx, "decode base64 fail", "err", err)
					return nil, 0, err
				}
				return b64, response.Usage.TotalTokens, nil
			}
		}
	}
	
	return nil, 0, errors.New("image is empty")
}

func GetMixClient(ctx context.Context, clientType string) *openrouter.Client {
	t := param.OpenRouter
	switch clientType {
	case "txt":
		t = utils.GetTxtType(db.GetCtxUserInfo(ctx).LLMConfigRaw)
	case "img":
		t = utils.GetImgType(db.GetCtxUserInfo(ctx).LLMConfigRaw)
	case "video":
		t = utils.GetVideoType(db.GetCtxUserInfo(ctx).LLMConfigRaw)
	case "rec":
		t = utils.GetRecType(db.GetCtxUserInfo(ctx).LLMConfigRaw)
	}
	
	token := ""
	specialLLMUrl := ""
	switch t {
	case param.OpenRouter:
		token = conf.BaseConfInfo.OpenRouterToken
	case param.AI302:
		token = conf.BaseConfInfo.AI302Token
		specialLLMUrl = "https://api.302.ai/"
	}
	
	config := openrouter.DefaultConfig(token)
	config.HTTPClient = utils.GetLLMProxyClient()
	if specialLLMUrl != "" {
		config.BaseURL = specialLLMUrl
	}
	if conf.BaseConfInfo.CustomUrl != "" {
		config.BaseURL = conf.BaseConfInfo.CustomUrl
	}
	return openrouter.NewClientWithConfig(*config)
}

func Generate302AIVideo(ctx context.Context, prompt string, image []byte) (string, int, error) {
	httpClient := utils.GetLLMProxyClient()
	
	start := time.Now()
	llmConfig := db.GetCtxUserInfo(ctx).LLMConfigRaw
	mediaType := utils.GetVideoType(llmConfig)
	model := utils.GetUsingVideoModel(mediaType, llmConfig.VideoModel)
	metrics.APIRequestCount.WithLabelValues(model).Inc()
	
	// Step 1: prepare payload using map -> json
	payloadMap := map[string]interface{}{
		"model":      model,
		"prompt":     prompt,
		"duration":   conf.VideoConfInfo.Duration,
		"resolution": conf.VideoConfInfo.Resolution,
		"fps":        conf.VideoConfInfo.FPS,
	}
	
	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal payload: %w", err)
	}
	payload := strings.NewReader(string(payloadBytes))
	
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.302.ai/302/v2/video/create", payload)
	
	metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())
	
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+conf.BaseConfInfo.AI302Token)
	req.Header.Add("Content-Type", "application/json")
	
	var res *http.Response
	for i := 0; i < conf.BaseConfInfo.LLMRetryTimes; i++ {
		res, err = httpClient.Do(req)
		if err != nil {
			time.Sleep(time.Duration(conf.BaseConfInfo.LLMRetryInterval) * time.Millisecond)
			continue
		}
		break
	}
	
	if err != nil || res == nil {
		return "", 0, fmt.Errorf("failed to call create API: %w %v", err, res)
	}
	defer res.Body.Close()
	
	body, _ := io.ReadAll(res.Body)
	
	var createResp CreateResp
	if err := json.Unmarshal(body, &createResp); err != nil {
		return "", 0, fmt.Errorf("failed to parse create response: %w, body=%s", err, string(body))
	}
	if createResp.TaskID == "" {
		return "", 0, fmt.Errorf("no task_id returned from create API, body=%s", string(body))
	}
	
	// Step 2: Poll fetch API (保持原逻辑)
	fetchURL := "https://api.302.ai/302/v2/video/fetch/" + createResp.TaskID
	for i := 0; i < 100; i++ {
		select {
		case <-ctx.Done():
			return "", 0, fmt.Errorf("context canceled or timeout: %w", ctx.Err())
		default:
		}
		
		req, _ := http.NewRequestWithContext(ctx, "GET", fetchURL, nil)
		req.Header.Add("Authorization", "Bearer "+conf.BaseConfInfo.AI302Token)
		
		res, err := httpClient.Do(req)
		if err != nil {
			logger.ErrorCtx(ctx, "failed to fetch result:", "err", err)
			time.Sleep(5 * time.Second)
			continue
		}
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		
		var fetchResp AI302FetchResp
		if err := json.Unmarshal(body, &fetchResp); err != nil {
			logger.ErrorCtx(ctx, "failed to parse fetch response:", "err", err, "body", string(body))
			time.Sleep(5 * time.Second)
			continue
		}
		
		if fetchResp.Status == "completed" {
			if fetchResp.VideoURL != "" {
				return fetchResp.VideoURL, 0, nil
			}
			return "", 0, fmt.Errorf("task completed but no video url found, body=%s", string(body))
		} else if fetchResp.Status == "failed" {
			return "", 0, fmt.Errorf("video generation failed: body=%s", string(body))
		} else {
			logger.InfoCtx(ctx, "task is still running, polling again...")
		}
		
		time.Sleep(5 * time.Second)
	}
	
	return "", 0, fmt.Errorf("video generation timeout")
}
