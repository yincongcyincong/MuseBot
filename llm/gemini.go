package llm

import (
	"context"
	"encoding/base64"
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
	"google.golang.org/genai/interactions/models/interactions"
	"google.golang.org/genai/interactions/models/operations"
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
	response, err = retryRequest(ctx, func() (*genai.GenerateContentResponse, error) {
		resp, err := client.Models.GenerateContent(
			ctx,
			model,
			geminiContent,
			&genai.GenerateContentConfig{
				ResponseModalities: []string{"TEXT", "IMAGE"},
			},
		)
		if err != nil {
			return nil, err
		}
		if resp == nil {
			return nil, fmt.Errorf("empty generate content response")
		}
		return resp, nil
	})

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

	// gemini-omni video models serve through the synchronous Interactions API
	// instead of the retired Models.GenerateVideos path; image-to-video sends
	// the image first, then the text prompt
	input := interactions.NewInteractionsInput(prompt)
	if len(image) > 0 {
		input = interactions.NewInteractionsInput([]interactions.Content{
			interactions.NewContent(interactions.ImageContent{
				Data:     genai.Ptr(base64.StdEncoding.EncodeToString(image)),
				MimeType: interactions.ImageContentMimeType("image/" + utils.DetectImageFormat(image)).ToPointer(),
			}),
			interactions.NewContent(interactions.TextContent{Text: prompt}),
		})
	}

	var res *operations.CreateInteractionResponse
	res, err = retryRequest(ctx, func() (*operations.CreateInteractionResponse, error) {
		resp, err := client.Interactions.Create(ctx, operations.CreateInteractionRequest{
			Body: operations.NewCreateInteractionRequestBody(interactions.CreateModelInteraction{
				Model:      interactions.Model(model),
				Input:      &input,
				Stream:     genai.Ptr(false),
				Store:      genai.Ptr(false),
				Background: genai.Ptr(false),
			}),
		})
		if err != nil {
			return nil, err
		}
		if resp == nil || resp.Interaction == nil {
			return nil, errors.New("empty interaction response")
		}
		return resp, nil
	})

	if err != nil || res == nil {
		logger.ErrorCtx(ctx, "generate video fail", "err", err)
		return nil, 0, err
	}

	metrics.APIRequestDuration.WithLabelValues(model).Observe(time.Since(start).Seconds())

	video := res.Interaction.OutputVideo
	if video == nil {
		logger.ErrorCtx(ctx, "generate video fail", "err", "video is empty", "interaction_id", res.Interaction.ID)
		return nil, 0, errors.New("video is empty")
	}

	var totalToken int
	if res.Interaction.Usage != nil && res.Interaction.Usage.TotalTokens != nil {
		totalToken = *res.Interaction.Usage.TotalTokens
	}

	// the video comes back inline as base64; oversized ones fall back to a file URI
	if video.Data != nil && *video.Data != "" {
		videoBytes, err := base64.StdEncoding.DecodeString(*video.Data)
		if err != nil {
			logger.ErrorCtx(ctx, "decode video fail", "err", err)
			return nil, 0, err
		}
		return videoBytes, totalToken, nil
	}

	if video.URI != nil && *video.URI != "" {
		videoBytes, err := downloadInteractionFile(ctx, client, *video.URI)
		if err != nil {
			logger.ErrorCtx(ctx, "download video fail", "err", err, "uri", *video.URI)
			return nil, 0, err
		}
		return videoBytes, totalToken, nil
	}

	logger.ErrorCtx(ctx, "generate video fail", "err", "video is empty", "interaction_id", res.Interaction.ID)
	return nil, 0, errors.New("video is empty")
}

// downloadInteractionFile fetches a video delivered as a Files API URI: poll
// the file until it is ACTIVE, then download its bytes.
func downloadInteractionFile(ctx context.Context, client *genai.Client, uri string) ([]byte, error) {
	fileName := uri
	if idx := strings.LastIndex(fileName, "/"); idx >= 0 {
		fileName = fileName[idx+1:]
	}
	if idx := strings.IndexAny(fileName, ":?"); idx >= 0 {
		fileName = fileName[:idx]
	}

	for i := 0; ; i++ {
		fileInfo, err := client.Files.Get(ctx, "files/"+fileName, nil)
		if err != nil {
			return nil, err
		}
		if fileInfo == nil {
			return nil, errors.New("video file is empty")
		}
		if fileInfo.State == genai.FileStateActive {
			return client.Files.Download(ctx, genai.NewDownloadURIFromFile(fileInfo), nil)
		}
		if fileInfo.State == genai.FileStateFailed {
			return nil, errors.New("video file processing failed")
		}
		if i >= 10 {
			return nil, errors.New("video file is not ready")
		}
		logger.InfoCtx(ctx, "video file is processing...")
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
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
	result, err = retryRequest(ctx, func() (*genai.GenerateContentResponse, error) {
		resp, err := client.Models.GenerateContent(
			ctx,
			model,
			contents,
			nil,
		)
		if err != nil {
			return nil, err
		}
		if resp == nil {
			return nil, fmt.Errorf("empty generate content response")
		}
		return resp, nil
	})

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
	response, err = retryRequest(ctx, func() (*genai.GenerateContentResponse, error) {
		resp, err := client.Models.GenerateContent(
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
			return nil, err
		}
		if resp == nil {
			return nil, fmt.Errorf("empty generate content response")
		}
		return resp, nil
	})

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
		// Trim with a cutset would strip every trailing '/', 'v' and '1' char
		httpOption.BaseURL = strings.TrimSuffix(conf.BaseConfInfo.CustomUrl, "/v1")
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

func init() {
	RegisterRecognizer(param.Gemini, GenerateGeminiText)
}
