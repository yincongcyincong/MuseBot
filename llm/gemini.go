package llm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/db"
	"github.com/yincongcyincong/MuseBot/i18n"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/metrics"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
	"google.golang.org/genai"
)

type GeminiReq struct {
	ToolCall           []*genai.FunctionCall
	ToolMessage        []*genai.Content
	CurrentToolMessage []*genai.Content
	
	GeminiMsgs []*genai.Content
}

func GenerateGeminiImg(ctx context.Context, prompt string, imageContent []byte) ([]byte, int, error) {
	client, err := GetGeminiClient(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "create client fail", "err", err)
		return nil, 0, err
	}
	
	start := time.Now()
	model := utils.GetUsingImgModel(param.Gemini, db.GetCtxUserInfo(ctx).LLMConfigRaw.ImgModel)
	metrics.APIRequestCount.WithLabelValues(model).Inc()
	
	geminiContent := genai.Text(prompt)
	if len(imageContent) > 0 {
		geminiContent = append(geminiContent, &genai.Content{
			Role: genai.RoleUser,
			Parts: []*genai.Part{
				{
					InlineData: &genai.Blob{
						Data:     imageContent,
						MIMEType: "image/" + utils.DetectImageFormat(imageContent),
					},
				},
			},
		})
	}
	
	var response *genai.GenerateContentResponse
	for i := 0; i < conf.BaseConfInfo.LLMRetryTimes; i++ {
		response, err = client.Models.GenerateContent(
			ctx,
			model,
			geminiContent,
			&genai.GenerateContentConfig{
				ResponseModalities: []string{"TEXT", "IMAGE"},
			},
		)
		
		if err != nil {
			logger.ErrorCtx(ctx, "generate image fail", "err", err)
			continue
		}
		break
	}
	
	if err != nil || response == nil {
		logger.ErrorCtx(ctx, "generate image fail", "err", err)
		return nil, 0, fmt.Errorf("request fail %v %v", err, response)
	}
	metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())
	
	if len(response.Candidates) > 0 && response.Candidates[0].Content != nil {
		for _, part := range response.Candidates[0].Content.Parts {
			if part.InlineData != nil {
				return part.InlineData.Data, int(response.UsageMetadata.TotalTokenCount), nil
			}
		}
	}
	
	return nil, 0, errors.New("image is empty")
}

func GenerateGeminiVideo(ctx context.Context, prompt string, image []byte) ([]byte, int, error) {
	client, err := GetGeminiClient(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "create client fail", "err", err)
		return nil, 0, err
	}
	
	start := time.Now()
	model := utils.GetUsingVideoModel(param.Gemini, db.GetCtxUserInfo(ctx).LLMConfigRaw.VideoModel)
	metrics.APIRequestCount.WithLabelValues(model).Inc()
	
	var geminiImage *genai.Image
	if len(image) > 0 {
		geminiImage = &genai.Image{
			ImageBytes: image,
			MIMEType:   "image/" + utils.DetectImageFormat(image),
		}
	}
	
	var operation *genai.GenerateVideosOperation
	for i := 0; i < conf.BaseConfInfo.LLMRetryTimes; i++ {
		operation, err = client.Models.GenerateVideos(ctx,
			model, prompt,
			geminiImage,
			&genai.GenerateVideosConfig{})
		if err != nil {
			logger.ErrorCtx(ctx, "generate video fail", "err", err)
			continue
		}
		break
	}
	
	if err != nil || operation == nil {
		logger.ErrorCtx(ctx, "generate video fail", "err", err, "operation", operation)
		return nil, 0, err
	}
	
	for !operation.Done {
		logger.InfoCtx(ctx, "video is createing...")
		time.Sleep(5 * time.Second)
		operation, err = client.Operations.GetVideosOperation(ctx, operation, nil)
		if err != nil {
			logger.ErrorCtx(ctx, "get video operation fail", "err", err)
			return nil, 0, err
		}
	}
	
	metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())
	
	if len(operation.Response.GeneratedVideos) == 0 {
		logger.ErrorCtx(ctx, "generate video fail", "err", "video is empty", "resp", operation.Response)
		return nil, 0, errors.New("video is empty")
	}
	
	var totalToken int
	if operation.Metadata != nil {
		if usageRaw, ok := operation.Metadata["usageMetadata"]; ok {
			if usage, ok := usageRaw.(map[string]interface{}); ok {
				if tokenValue, ok := usage["totalTokenCount"]; ok {
					if tokenFloat, ok := tokenValue.(float64); ok {
						totalToken = int(tokenFloat)
					}
				}
			}
		}
	}
	
	return operation.Response.GeneratedVideos[0].Video.VideoBytes, totalToken, nil
}

func GenerateGeminiText(ctx context.Context, audioContent []byte) (string, int, error) {
	client, err := GetGeminiClient(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "create client fail", "err", err)
		return "", 0, err
	}
	
	start := time.Now()
	model := utils.GetUsingRecModel(param.Gemini, db.GetCtxUserInfo(ctx).LLMConfigRaw.RecModel)
	metrics.APIRequestCount.WithLabelValues(model).Inc()
	
	parts := []*genai.Part{
		genai.NewPartFromText("Get Content from this audio clip"),
		{
			InlineData: &genai.Blob{
				MIMEType: "audio/" + utils.DetectAudioFormat(audioContent),
				Data:     audioContent,
			},
		},
	}
	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}
	
	var result *genai.GenerateContentResponse
	for i := 0; i < conf.BaseConfInfo.LLMRetryTimes; i++ {
		result, err = client.Models.GenerateContent(
			ctx,
			model,
			contents,
			nil,
		)
		
		if err != nil || result == nil {
			logger.ErrorCtx(ctx, "generate text fail", "err", err)
			continue
		}
		break
	}
	
	if err != nil || result == nil {
		logger.ErrorCtx(ctx, "generate text fail", "err", err)
		return "", 0, err
	}
	
	metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())
	return result.Text(), int(result.UsageMetadata.TotalTokenCount), nil
}

func GeminiTTS(ctx context.Context, content, encoding string) ([]byte, int, int, error) {
	client, err := GetGeminiClient(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "create client fail", "err", err)
		return nil, 0, 0, err
	}
	
	start := time.Now()
	model := utils.GetUsingTTSModel(param.Gemini, db.GetCtxUserInfo(ctx).LLMConfigRaw.TTSModel)
	metrics.APIRequestCount.WithLabelValues(model).Inc()
	
	parts := []*genai.Part{
		genai.NewPartFromText(i18n.GetMessage("audio_create_prompt", map[string]interface{}{
			"content": content,
		})),
	}
	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}
	
	var response *genai.GenerateContentResponse
	for i := 0; i < conf.BaseConfInfo.LLMRetryTimes; i++ {
		response, err = client.Models.GenerateContent(
			ctx,
			model,
			contents,
			&genai.GenerateContentConfig{
				ResponseModalities: []string{
					"AUDIO",
				},
				SpeechConfig: &genai.SpeechConfig{
					VoiceConfig: &genai.VoiceConfig{
						PrebuiltVoiceConfig: &genai.PrebuiltVoiceConfig{
							VoiceName: conf.AudioConfInfo.GeminiVoiceName,
						},
					},
				},
			},
		)
		
		if err != nil {
			logger.ErrorCtx(ctx, "generate audio fail", "err", err)
			continue
		}
		break
	}
	
	if err != nil || response == nil {
		logger.ErrorCtx(ctx, "generate audio fail", "err", err)
		return nil, 0, 0, fmt.Errorf("request fail %v %v", err, response)
	}
	
	metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())
	if len(response.Candidates) > 0 {
		for _, part := range response.Candidates[0].Content.Parts {
			if part.InlineData != nil {
				var data = part.InlineData.Data
				data, err = utils.GetAudioData(encoding, part.InlineData.Data)
				if err != nil {
					logger.ErrorCtx(ctx, "convert audio fail", "err", err)
				}
				return data, int(response.UsageMetadata.TotalTokenCount), utils.PCMDuration(len(part.InlineData.Data), 24000, 1, 16), nil
			}
		}
	}
	
	return nil, 0, 0, errors.New("audio is empty")
}

func GetGeminiClient(ctx context.Context) (*genai.Client, error) {
	httpClient := utils.GetLLMProxyClient()
	httpOption := genai.HTTPOptions{}
	if conf.BaseConfInfo.CustomUrl != "" {
		httpOption.BaseURL = strings.Trim(conf.BaseConfInfo.CustomUrl, "/v1")
		httpOption.Headers = http.Header{
			"Authorization": []string{"Bearer " + conf.BaseConfInfo.GeminiToken},
		}
	}
	return genai.NewClient(ctx, &genai.ClientConfig{
		HTTPClient:  httpClient,
		APIKey:      conf.BaseConfInfo.GeminiToken,
		HTTPOptions: httpOption,
	})
}
