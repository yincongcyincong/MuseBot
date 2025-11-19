package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

type MessageContent struct {
	Image string `json:"image,omitempty"`
	Text  string `json:"text,omitempty"`
}

type Message struct {
	Role    string            `json:"role"`
	Content []*MessageContent `json:"content"`
}

type Input struct {
	Messages []*Message `json:"messages"`
}

type Parameters struct {
	NegativePrompt string `json:"negative_prompt"`
	PromptExtend   bool   `json:"prompt_extend"`
	Watermark      bool   `json:"watermark"`
	Size           string `json:"size"`
}

type Payload struct {
	Model      string     `json:"model"`
	Input      Input      `json:"input"`
	Parameters Parameters `json:"parameters"`
}

type ImageResponse struct {
	StatusCode string `json:"status_code"`
	Message    string `json:"message"`
	Output     struct {
		Choices []struct {
			Message struct {
				Content []struct {
					Image string `json:"image"`
				} `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	} `json:"output"`
}

type TextResponse struct {
	StatusCode string `json:"status_code"`
	Message    string `json:"message"`
	Output     struct {
		Choices []struct {
			Message struct {
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	} `json:"output"`
	Usage struct {
		AudioTokens  int `json:"audio_tokens"`
		OutputTokens int `json:"output_tokens"`
		InputTokens  int `json:"input_tokens"`
	}
}

func GenerateAliyunImg(ctx context.Context, prompt string, imageContent []byte) (string, int, error) {
	url := "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation"
	
	model := utils.GetUsingImgModel(param.Aliyun, db.GetCtxUserInfo(ctx).LLMConfigRaw.ImgModel)
	if imageContent != nil {
		model = "qwen-image-edit"
	}
	
	payload := Payload{
		Model: model,
		Input: Input{
			Messages: []*Message{
				{
					Role: "user",
					Content: []*MessageContent{
						{Text: prompt},
					},
				},
			},
		},
		Parameters: Parameters{
			NegativePrompt: "",
			PromptExtend:   true,
			Watermark:      false,
			Size:           "1328*1328",
		},
	}
	
	if len(imageContent) > 0 {
		payload.Input.Messages[0].Content = append(payload.Input.Messages[0].Content,
			&MessageContent{Image: fmt.Sprintf("data:image/%s;base64,%s", utils.DetectImageFormat(imageContent), base64.StdEncoding.EncodeToString(imageContent)),
			})
	}
	
	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+*conf.BaseConfInfo.AliyunToken)
	
	client := utils.GetLLMProxyClient()
	start := time.Now()
	metrics.APIRequestCount.WithLabelValues(model).Inc()
	
	var resp *http.Response
	var err error
	for i := 0; i < *conf.BaseConfInfo.LLMRetryTimes; i++ {
		resp, err = client.Do(req)
		if err != nil {
			logger.ErrorCtx(ctx, "create image fail", "err", err)
			continue
		}
		break
	}
	if err != nil || resp == nil {
		return "", 0, fmt.Errorf("request fail %v %v", err, resp)
	}
	
	defer resp.Body.Close()
	metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, "read response fail", "err", err)
		return "", 0, err
	}
	
	var imgRsp ImageResponse
	if err := json.Unmarshal(body, &imgRsp); err != nil {
		logger.ErrorCtx(ctx, "unmarshal response fail", "err", err)
		return "", 0, err
	}
	
	if len(imgRsp.Output.Choices) == 0 {
		logger.ErrorCtx(ctx, "generate image fail", "message", imgRsp.Message)
		return "", 0, fmt.Errorf("generate image fail: %s", imgRsp.Message)
	}
	
	return imgRsp.Output.Choices[0].Message.Content[0].Image, param.ImageTokenUsage, nil
}

type videoRequest struct {
	Model      string                 `json:"model"`
	Input      map[string]interface{} `json:"input"`
	Parameters map[string]interface{} `json:"parameters"`
}

type TaskStatusResponse struct {
	RequestID string `json:"request_id"`
	Output    struct {
		TaskStatus string `json:"task_status"`
		TaskID     string `json:"task_id"`
	} `json:"output"`
}

type videoResponse struct {
	RequestID string `json:"request_id"`
	Output    struct {
		TaskID       string `json:"task_id"`
		TaskStatus   string `json:"task_status"`
		VideoURL     string `json:"video_url"`
		OrigPrompt   string `json:"orig_prompt"`
		ActualPrompt string `json:"actual_prompt"`
	} `json:"output"`
}

func GenerateAliyunVideo(ctx context.Context, prompt string, image []byte) (string, int, error) {
	
	input := map[string]interface{}{
		"prompt": prompt,
	}
	
	model := utils.GetUsingVideoModel(param.Aliyun, db.GetCtxUserInfo(ctx).LLMConfigRaw.VideoModel)
	if len(image) > 0 {
		base64Img := base64.StdEncoding.EncodeToString(image)
		input["img_url"] = fmt.Sprintf("data:image/%s;base64,%s", utils.DetectImageFormat(image), base64Img)
	}
	
	reqBody := videoRequest{
		Model: model,
		Input: input,
		Parameters: map[string]interface{}{
			"duration":      5,
			"audio":         true,
			"prompt_extend": true,
			"size":          "832*480",
			"resolution":    "480P",
		},
	}
	
	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, err
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", "https://dashscope.aliyuncs.com/api/v1/services/aigc/video-generation/video-synthesis", bytes.NewReader(data))
	if err != nil {
		return "", 0, err
	}
	
	req.Header.Set("Authorization", "Bearer "+*conf.BaseConfInfo.AliyunToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-DashScope-Async", "enable")
	
	client := utils.GetLLMProxyClient()
	start := time.Now()
	metrics.APIRequestCount.WithLabelValues(model).Inc()
	defer func() {
		metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())
	}()
	
	var resp *http.Response
	for i := 0; i < *conf.BaseConfInfo.LLMRetryTimes; i++ {
		resp, err = client.Do(req)
		if err != nil {
			logger.ErrorCtx(ctx, "create video fail", "err", err)
			continue
		}
		break
	}
	
	if err != nil || resp == nil {
		return "", 0, fmt.Errorf("request fail %v %v", err, resp)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}
	
	var vr TaskStatusResponse
	if err = json.Unmarshal(body, &vr); err != nil {
		return "", 0, err
	}
	
	for i := 0; i < 100; i++ {
		time.Sleep(5 * time.Second)
		
		req, err = http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://dashscope.aliyuncs.com/api/v1/tasks/%s", vr.Output.TaskID), nil)
		if err != nil {
			return "", 0, err
		}
		
		req.Header.Set("Authorization", "Bearer "+*conf.BaseConfInfo.AliyunToken)
		req.Header.Set("Content-Type", "application/json")
		
		resp, err = client.Do(req)
		if err != nil {
			return "", 0, err
		}
		defer resp.Body.Close()
		
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return "", 0, err
		}
		
		var vr videoResponse
		if err = json.Unmarshal(body, &vr); err != nil {
			return "", 0, err
		}
		
		if vr.Output.TaskStatus == "SUCCEEDED" && vr.Output.VideoURL != "" {
			return vr.Output.VideoURL, param.VideoTokenUsage, nil
		}
		
		if vr.Output.TaskStatus == "RUNNING" {
			logger.InfoCtx(ctx, "video is createing...")
		}
	}
	
	return "", 0, fmt.Errorf("video generation timeout")
}

func GenerateAliyunText(ctx context.Context, audioContent []byte) (string, int, error) {
	audioType := utils.DetectAudioFormat(audioContent)
	var err error
	switch audioType {
	case "mp4":
		audioType = "mp3"
		audioContent, err = utils.MP4ToMP3(audioContent)
		if err != nil {
			return "", 0, err
		}
	case "ogg":
		audioType = "mp3"
		audioContent, err = utils.OGGToMP3(audioContent)
		if err != nil {
			return "", 0, err
		}
	}
	
	audioBase64 := base64.StdEncoding.EncodeToString(audioContent)
	audioDataURL := fmt.Sprintf("data:audio/%s;base64,%s", audioType, audioBase64)
	
	recModel := utils.GetUsingRecModel(param.Aliyun, db.GetCtxUserInfo(ctx).LLMConfigRaw.RecModel)
	payload := map[string]interface{}{
		"model": "qwen-audio-turbo-latest",
		"input": map[string]interface{}{
			"messages": []map[string]interface{}{
				{
					"role": "user",
					"content": []map[string]interface{}{
						{"audio": audioDataURL},
						{"text": i18n.GetMessage("audio_rec_prompt", nil)},
					},
				},
			},
		},
	}
	
	data, err := json.Marshal(payload)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation",
		bytes.NewReader(data))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+*conf.BaseConfInfo.AliyunToken)
	req.Header.Set("Content-Type", "application/json")
	
	client := utils.GetLLMProxyClient()
	start := time.Now()
	metrics.APIRequestCount.WithLabelValues(recModel).Inc()
	
	var resp *http.Response
	for i := 0; i < *conf.BaseConfInfo.LLMRetryTimes; i++ {
		resp, err = client.Do(req)
		if err != nil {
			logger.ErrorCtx(ctx, "create video fail", "err", err)
			continue
		}
		break
	}
	
	if err != nil || resp == nil {
		return "", 0, fmt.Errorf("request fail %v %v", err, resp)
	}
	metrics.APIRequestDuration.WithLabelValues(recModel).Observe(time.Since(start).Seconds())
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
	}
	
	var tr TextResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", resp.StatusCode, fmt.Errorf("failed to parse response: %w", err)
	}
	
	if len(tr.Output.Choices) > 0 &&
		len(tr.Output.Choices[0].Message.Content) > 0 &&
		tr.Output.Choices[0].Message.Content[0].Text != "" {
		return tr.Output.Choices[0].Message.Content[0].Text, tr.Usage.OutputTokens + tr.Usage.InputTokens + tr.Usage.AudioTokens, nil
	}
	
	return "", resp.StatusCode, fmt.Errorf("no text returned, message: %s", tr.Message)
}

type TTSResponse struct {
	StatusCode int    `json:"status_code"`
	RequestID  string `json:"request_id"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Output     struct {
		Audio struct {
			Data      string `json:"data"`
			URL       string `json:"url"`
			ID        string `json:"id"`
			ExpiresAt int64  `json:"expires_at"`
		} `json:"audio"`
	} `json:"output"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
		Characters  int `json:"characters"`
	} `json:"usage"`
}

func AliyunTTS(ctx context.Context, text, encoding string) ([]byte, int, int, error) {
	url := "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation"
	
	model := utils.GetUsingTTSModel(param.Aliyun, db.GetCtxUserInfo(ctx).LLMConfigRaw.TTSModel)
	payload := map[string]interface{}{
		"model": model,
		"input": map[string]interface{}{
			"text":          text,
			"voice":         *conf.AudioConfInfo.AliyunAudioVoice,
			"language_type": "Auto",
		},
	}
	
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("marshal payload error: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, 0, fmt.Errorf("new request error: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+*conf.BaseConfInfo.AliyunToken)
	req.Header.Set("Content-Type", "application/json")
	
	client := utils.GetLLMProxyClient()
	start := time.Now()
	metrics.APIRequestCount.WithLabelValues(model).Inc()
	
	var resp *http.Response
	for i := 0; i < *conf.BaseConfInfo.LLMRetryTimes; i++ {
		resp, err = client.Do(req)
		if err != nil {
			logger.ErrorCtx(ctx, "create video fail", "err", err)
			continue
		}
		break
	}
	
	if err != nil || resp == nil {
		return nil, 0, 0, fmt.Errorf("request fail %v %v", err, resp)
	}
	
	metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())
	defer resp.Body.Close()
	
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("read response error: %w", err)
	}
	
	var ttsResp TTSResponse
	if err := json.Unmarshal(respData, &ttsResp); err != nil {
		return nil, 0, 0, fmt.Errorf("unmarshal response error: %w", err)
	}
	
	if ttsResp.Output.Audio.URL == "" {
		return nil, 0, 0, fmt.Errorf("API error: %+v", ttsResp)
	}
	
	audioContent, err := utils.DownloadFile(ttsResp.Output.Audio.URL)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("download audio error: %w", err)
	}
	
	audioContent, err = utils.WavToPCMBytes(audioContent)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("convert wav to pcm error: %w", err)
	}
	
	audioContent, err = utils.GetAudioDataDetail(encoding, audioContent, 16000, 1)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("get audio data error: %w", err)
	}
	
	return audioContent, param.AudioTokenUsage, utils.PCMDuration(len(audioContent), 16000, 1, 16), nil
}
