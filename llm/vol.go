package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
	
	"github.com/google/uuid"
	"github.com/volcengine/volc-sdk-golang/service/visual"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

// GenerateVolImg generate image
func GenerateVolImg(ctx context.Context, prompt string, imageContent []byte) (string, int, error) {
	start := time.Now()
	metrics.APIRequestCount.WithLabelValues(*conf.PhotoConfInfo.ModelVersion).Inc()
	
	visual.DefaultInstance.Client.SetAccessKey(*conf.BaseConfInfo.VolcAK)
	visual.DefaultInstance.Client.SetSecretKey(*conf.BaseConfInfo.VolcSK)
	
	reqBody := map[string]interface{}{
		"req_key":           *conf.PhotoConfInfo.ReqKey,
		"prompt":            prompt,
		"model_version":     *conf.PhotoConfInfo.ModelVersion,
		"req_schedule_conf": *conf.PhotoConfInfo.ReqScheduleConf,
		"llm_seed":          *conf.PhotoConfInfo.Seed,
		"seed":              *conf.PhotoConfInfo.Seed,
		"scale":             *conf.PhotoConfInfo.Scale,
		"ddim_steps":        *conf.PhotoConfInfo.DDIMSteps,
		"width":             *conf.PhotoConfInfo.Width,
		"height":            *conf.PhotoConfInfo.Height,
		"use_pre_llm":       *conf.PhotoConfInfo.UsePreLLM,
		"use_sr":            *conf.PhotoConfInfo.UseSr,
		"return_url":        *conf.PhotoConfInfo.ReturnUrl,
		"logo_info": map[string]interface{}{
			"add_logo":          *conf.PhotoConfInfo.AddLogo,
			"position":          *conf.PhotoConfInfo.Position,
			"language":          *conf.PhotoConfInfo.Language,
			"opacity":           *conf.PhotoConfInfo.Opacity,
			"logo_text_content": *conf.PhotoConfInfo.LogoTextContent,
		},
	}
	
	if len(imageContent) != 0 {
		reqBody["binary_data_base64"] = []string{base64.StdEncoding.EncodeToString(imageContent)}
	}
	
	resp, _, err := visual.DefaultInstance.CVProcess(reqBody)
	if err != nil {
		logger.ErrorCtx(ctx, "request img api fail", "err", err)
		return "", 0, err
	}
	
	respByte, _ := json.Marshal(resp)
	data := &param.ImgResponse{}
	err = json.Unmarshal(respByte, data)
	if err != nil {
		logger.ErrorCtx(ctx, "unmarshal response fail", "err", err)
		return "", 0, err
	}
	
	logger.InfoCtx(ctx, "image response", "respByte", respByte)
	
	metrics.APIRequestDuration.WithLabelValues(*conf.PhotoConfInfo.ModelVersion).Observe(time.Since(start).Seconds())
	
	if data.Data == nil || len(data.Data.ImageUrls) == 0 {
		logger.WarnCtx(ctx, "no image generated")
		return "", 0, errors.New("no image generated")
	}
	
	return data.Data.ImageUrls[0], param.ImageTokenUsage, nil
}

// GenerateVolVideo generate video
func GenerateVolVideo(ctx context.Context, prompt string, imageContent []byte) (string, int, error) {
	if prompt == "" {
		logger.WarnCtx(ctx, "prompt is empty", "prompt", prompt)
		return "", 0, errors.New("prompt is empty")
	}
	
	start := time.Now()
	metrics.APIRequestCount.WithLabelValues(*conf.PhotoConfInfo.ModelVersion).Inc()
	
	client := GetVolClient()
	videoParam := fmt.Sprintf(" --ratio %s --fps %d  --dur %d --resolution %s --watermark %t",
		*conf.VideoConfInfo.Radio, *conf.VideoConfInfo.FPS, *conf.VideoConfInfo.Duration, *conf.VideoConfInfo.Resolution, *conf.VideoConfInfo.Watermark)
	
	text := prompt + videoParam
	contents := make([]*model.CreateContentGenerationContentItem, 0)
	contents = append(contents, &model.CreateContentGenerationContentItem{
		Type: model.ContentGenerationContentItemTypeText,
		Text: &text,
	})
	
	if len(imageContent) > 0 {
		frame := "first_frame"
		contents = append(contents, &model.CreateContentGenerationContentItem{
			Type: model.ContentGenerationContentItemTypeImage,
			ImageURL: &model.ImageURL{
				URL: "data:image/" + utils.DetectImageFormat(imageContent) + ";base64," + base64.StdEncoding.EncodeToString(imageContent),
			},
			Role: &frame,
		})
	}
	
	llmConfig := db.GetCtxUserInfo(ctx).LLMConfigRaw
	mediaType := utils.GetVideoType(llmConfig)
	modelStr := utils.GetUsingVideoModel(mediaType, llmConfig.VideoModel)
	resp, err := client.CreateContentGenerationTask(ctx, model.CreateContentGenerationTaskRequest{
		Model:   modelStr,
		Content: contents,
	})
	if err != nil {
		logger.ErrorCtx(ctx, "request create video api fail", "err", err)
		return "", 0, err
	}
	
	metrics.APIRequestDuration.WithLabelValues(*conf.PhotoConfInfo.ModelVersion).Observe(time.Since(start).Seconds())
	for i := 0; i < 100; i++ {
		getResp, err := client.GetContentGenerationTask(ctx, model.GetContentGenerationTaskRequest{
			ID: resp.ID,
		})
		
		if err != nil {
			logger.ErrorCtx(ctx, "request get video api fail", "err", err)
			return "", 0, err
		}
		
		if getResp.Status == model.StatusRunning || getResp.Status == model.StatusQueued {
			logger.InfoCtx(ctx, "video is createing...")
			time.Sleep(5 * time.Second)
			continue
		}
		
		if getResp.Error != nil {
			logger.ErrorCtx(ctx, "request get video api fail", "err", getResp.Error)
			return "", 0, errors.New(getResp.Error.Message)
		}
		
		if getResp.Status == model.StatusSucceeded {
			return getResp.Content.VideoURL, getResp.Usage.TotalTokens, nil
		} else {
			logger.ErrorCtx(ctx, "request get video api fail", "status", getResp.Status)
			return "", 0, errors.New("create video fail")
		}
	}
	
	return "", 0, fmt.Errorf("video generation timeout")
}

func GetVolImageContent(ctx context.Context, imageContent []byte, content string) (string, int, error) {
	client := GetVolClient()
	start := time.Now()
	
	llmConfig := db.GetCtxUserInfo(ctx).LLMConfigRaw
	mediaType := utils.GetRecType(llmConfig)
	modelStr := utils.GetUsingRecModel(mediaType, llmConfig.RecModel)
	
	metrics.APIRequestCount.WithLabelValues(modelStr).Inc()
	
	contentPrompt := content
	if content == "" {
		contentPrompt = i18n.GetMessage(*conf.BaseConfInfo.Lang, "photo_handle_prompt", nil)
	}
	
	req := model.ChatCompletionRequest{
		Model: modelStr,
		Messages: []*model.ChatCompletionMessage{
			{
				Role: model.ChatMessageRoleUser,
				Content: &model.ChatCompletionMessageContent{
					ListValue: []*model.ChatCompletionMessageContentPart{
						{
							Type: model.ChatCompletionMessageContentPartTypeImageURL,
							ImageURL: &model.ChatMessageImageURL{
								URL: "data:image/" + utils.DetectImageFormat(imageContent) + ";base64," + base64.StdEncoding.EncodeToString(imageContent),
							},
						},
						{
							Type: model.ChatCompletionMessageContentPartTypeText,
							Text: contentPrompt,
						},
					},
				},
			},
		},
	}
	
	response, err := client.CreateChatCompletion(ctx, req)
	metrics.APIRequestDuration.WithLabelValues(modelStr).Observe(time.Since(start).Seconds())
	if err != nil {
		logger.ErrorCtx(ctx, "CreateChatCompletion error", "err", err)
		return "", 0, err
	}
	
	if len(response.Choices) == 0 {
		logger.ErrorCtx(ctx, "response is emtpy", "response", response)
		return "", 0, errors.New("response is empty")
	}
	
	return *response.Choices[0].Message.Content.StringValue, response.Usage.TotalTokens, nil
}

type TTSServResponse struct {
	ReqID     string `json:"reqid"`
	Code      int    `json:"code"`
	Message   string `json:"Message"`
	Operation string `json:"operation"`
	Sequence  int    `json:"sequence"`
	Data      string `json:"data"`
	Addition  struct {
		Duration string `json:"duration"`
	} `json:"addition"`
}

func VolTTS(ctx context.Context, text, userId, encoding string) ([]byte, int, int, error) {
	start := time.Now()
	
	model := utils.GetUsingTTSModel(param.Vol, db.GetCtxUserInfo(ctx).LLMConfigRaw.TTSModel)
	metrics.APIRequestCount.WithLabelValues(model).Inc()
	
	formatEncoding := encoding
	if encoding != "mp3" && encoding != "wav" && encoding != "ogg_opus" && encoding != "pcm" {
		formatEncoding = "pcm"
	}
	
	reqID := uuid.NewString()
	params := make(map[string]map[string]interface{})
	params["app"] = make(map[string]interface{})
	
	params["app"]["appid"] = *conf.AudioConfInfo.VolAudioAppID
	params["app"]["token"] = *conf.AudioConfInfo.VolAudioToken
	params["app"]["cluster"] = model
	params["user"] = make(map[string]interface{})
	
	params["user"]["uid"] = userId
	params["audio"] = make(map[string]interface{})
	
	params["audio"]["voice_type"] = *conf.AudioConfInfo.VolAudioVoiceType
	params["audio"]["encoding"] = formatEncoding
	params["audio"]["speed_ratio"] = 1.0
	params["audio"]["volume_ratio"] = 1.0
	params["audio"]["pitch_ratio"] = 1.0
	params["request"] = make(map[string]interface{})
	params["request"]["reqid"] = reqID
	params["request"]["text"] = text
	params["request"]["text_type"] = "plain"
	params["request"]["operation"] = "query"
	
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Authorization"] = fmt.Sprintf("Bearer;%s", *conf.AudioConfInfo.VolAudioToken)
	
	url := "https://openspeech.bytedance.com/api/v1/tts"
	bodyStr, _ := json.Marshal(params)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyStr))
	if err != nil {
		logger.ErrorCtx(ctx, "NewRequest error", "err", err)
		return nil, 0, 0, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	httpClient := utils.GetLLMProxyClient()
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.ErrorCtx(ctx, "httpClient.Do error", "err", err)
		return nil, 0, 0, err
	}
	
	metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())
	
	synResp, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorCtx(ctx, "io.ReadAll error", "err", err)
		return nil, 0, 0, err
	}
	
	var respJSON TTSServResponse
	err = json.Unmarshal(synResp, &respJSON)
	if err != nil {
		return nil, 0, 0, err
	}
	code := respJSON.Code
	if code != 3000 {
		logger.ErrorCtx(ctx, "resp code fail", "code", code, "message", respJSON.Message)
		return nil, 0, 0, errors.New("resp code fail")
	}
	
	audio, _ := base64.StdEncoding.DecodeString(respJSON.Data)
	if formatEncoding == "pcm" {
		audio, err = utils.GetAudioData(encoding, audio)
		if err != nil {
			logger.ErrorCtx(ctx, "EncodePcmBuffToSilk error", "err", err)
			return nil, 0, 0, err
		}
	}
	
	return audio, param.AudioTokenUsage, utils.ParseInt(respJSON.Addition.Duration), nil
}

func GetVolClient() *arkruntime.Client {
	httpClient := utils.GetLLMProxyClient()
	return arkruntime.NewClientWithApiKey(
		*conf.BaseConfInfo.VolToken,
		arkruntime.WithTimeout(5*time.Minute),
		arkruntime.WithHTTPClient(httpClient),
	)
}
