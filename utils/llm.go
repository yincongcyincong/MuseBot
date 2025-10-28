package utils

import (
	godeepseek "github.com/cohesion-org/deepseek-go"
	"github.com/devinyf/dashscopego/qwen"
	"github.com/goccy/go-json"
	"github.com/sashabaranov/go-openai"
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/param"
)

func GetDefaultLLMConfig() string {
	llmConf := param.LLMConfig{
		TxtType:    *conf.BaseConfInfo.Type,
		ImgType:    *conf.BaseConfInfo.MediaType,
		VideoType:  *conf.BaseConfInfo.MediaType,
		TxtModel:   GetTxtModel(*conf.BaseConfInfo.Type),
		ImgModel:   GetImgModel(*conf.BaseConfInfo.MediaType),
		VideoModel: GetVideoModel(*conf.BaseConfInfo.MediaType),
	}
	b, _ := json.Marshal(llmConf)
	return string(b)
}

func GetTxtType(llmConf *param.LLMConfig) string {
	if llmConf == nil {
		return *conf.BaseConfInfo.Type
	}
	
	aType := GetAvailTxtType()
	for _, v := range aType {
		if v == llmConf.TxtType {
			return v
		}
	}
	
	if len(aType) > 0 {
		return aType[0]
	}
	
	return *conf.BaseConfInfo.Type
}

func GetImgType(llmConf *param.LLMConfig) string {
	if llmConf == nil {
		return *conf.BaseConfInfo.MediaType
	}
	
	aType := GetAvailImgType()
	for _, v := range aType {
		if v == llmConf.ImgType {
			return v
		}
	}
	
	if len(aType) > 0 {
		return aType[0]
	}
	
	return *conf.BaseConfInfo.MediaType
}

func GetVideoType(llmConf *param.LLMConfig) string {
	if llmConf == nil {
		return *conf.BaseConfInfo.MediaType
	}
	
	aType := GetAvailVideoType()
	for _, v := range aType {
		if v == llmConf.VideoType {
			return v
		}
	}
	
	if len(aType) > 0 {
		return aType[0]
	}
	
	return *conf.BaseConfInfo.MediaType
}

func GetImgModel(t string) string {
	switch t {
	case param.Gemini:
		return *conf.PhotoConfInfo.GeminiImageModel
	case param.OpenAi:
		return *conf.PhotoConfInfo.OpenAIImageModel
	case param.Aliyun:
		return *conf.PhotoConfInfo.AliyunImageModel
	case param.AI302:
		return *conf.PhotoConfInfo.MixImageModel
	case param.Vol:
		return *conf.PhotoConfInfo.VolImageModel
	case param.ChatAnyWhere:
		return *conf.PhotoConfInfo.OpenAIImageModel
	}
	
	return ""
}

func GetVideoModel(t string) string {
	switch t {
	case param.Gemini:
		return *conf.VideoConfInfo.GeminiVideoModel
	case param.Aliyun:
		return *conf.VideoConfInfo.AliyunVideoModel
	case param.AI302:
		return *conf.VideoConfInfo.AI302VideoModel
	case param.Vol:
		return *conf.VideoConfInfo.VolVideoModel
	}
	
	return ""
}

func GetTxtModel(t string) string {
	switch t {
	case param.DeepSeek:
		return godeepseek.DeepSeekChat
	case param.Gemini:
		return param.ModelGemini20Flash
	case param.OpenAi:
		return openai.GPT3Dot5Turbo
	case param.OpenRouter:
		return param.DeepseekDeepseekR1_0528Free
	case param.AI302:
		return param.DeepseekDeepseekR1_0528
	case param.Ollama:
		return "deepseek-r1"
	case param.Vol:
		return param.ModelDeepSeekR1_528
	case param.Aliyun:
		return qwen.QwenMax
	case param.ChatAnyWhere:
		return openai.GPT3Dot5Turbo
	}
	
	return godeepseek.DeepSeekChat
}

func GetAvailTxtType() []string {
	res := []string{}
	if *conf.BaseConfInfo.DeepseekToken != "" {
		res = append(res, param.DeepSeek)
	}
	if *conf.BaseConfInfo.GeminiToken != "" {
		res = append(res, param.Gemini)
	}
	if *conf.BaseConfInfo.OpenAIToken != "" {
		res = append(res, param.OpenAi)
	}
	if *conf.BaseConfInfo.AliyunToken != "" {
		res = append(res, param.Aliyun)
	}
	if *conf.BaseConfInfo.VolToken != "" {
		res = append(res, param.Vol)
	}
	if *conf.BaseConfInfo.ChatAnyWhereToken != "" {
		res = append(res, param.ChatAnyWhere)
	}
	if *conf.BaseConfInfo.AI302Token != "" {
		res = append(res, param.AI302)
	}
	if *conf.BaseConfInfo.OpenRouterToken != "" {
		res = append(res, param.OpenRouter)
	}
	if *conf.BaseConfInfo.Type == param.Ollama {
		res = append(res, param.Ollama)
	}
	return res
}

func GetAvailImgType() []string {
	res := []string{}
	
	if *conf.BaseConfInfo.GeminiToken != "" {
		res = append(res, param.Gemini)
	}
	if *conf.BaseConfInfo.OpenRouterToken != "" {
		res = append(res, param.OpenAi)
	}
	if *conf.BaseConfInfo.AliyunToken != "" {
		res = append(res, param.Aliyun)
	}
	if *conf.BaseConfInfo.AI302Token != "" {
		res = append(res, param.AI302)
	}
	if *conf.BaseConfInfo.VolToken != "" {
		res = append(res, param.Vol)
	}
	if *conf.BaseConfInfo.ChatAnyWhereToken != "" {
		res = append(res, param.ChatAnyWhere)
	}
	
	return res
}

func GetAvailVideoType() []string {
	res := []string{}
	if *conf.BaseConfInfo.GeminiToken != "" {
		res = append(res, param.Gemini)
	}
	if *conf.BaseConfInfo.AliyunToken != "" {
		res = append(res, param.Aliyun)
	}
	if *conf.BaseConfInfo.AI302Token != "" {
		res = append(res, param.AI302)
	}
	if *conf.BaseConfInfo.VolToken != "" {
		res = append(res, param.Vol)
	}
	
	return res
}

func GetAvailRecType() []string {
	res := []string{}
	
	if *conf.BaseConfInfo.GeminiToken != "" {
		res = append(res, param.Gemini)
	}
	if *conf.BaseConfInfo.OpenRouterToken != "" {
		res = append(res, param.OpenAi)
	}
	if *conf.BaseConfInfo.AliyunToken != "" {
		res = append(res, param.Aliyun)
	}
	if *conf.BaseConfInfo.AI302Token != "" {
		res = append(res, param.AI302)
	}
	if *conf.BaseConfInfo.VolToken != "" {
		res = append(res, param.Vol)
	}
	
	return res
}

func GetRecType(llmConf *param.LLMConfig) string {
	if llmConf == nil {
		return *conf.BaseConfInfo.MediaType
	}
	
	aType := GetAvailRecType()
	for _, v := range aType {
		if v == llmConf.RecType {
			return v
		}
	}
	
	if len(aType) > 0 {
		return aType[0]
	}
	
	return *conf.BaseConfInfo.MediaType
}

func GetUsingImgModel(ty string, model string) string {
	switch ty {
	case param.Gemini:
		if param.GeminiImageModels[model] {
			return model
		}
		return param.GeminiImageGenPreview
	
	case param.OpenAi:
		if param.OpenAIImageModels[model] {
			return model
		}
		return param.ModelImageGPT
	case param.Aliyun:
		if param.AliyunImageModels[model] {
			return model
		}
		return param.QwenImagePlus
	case param.AI302:
		return model
	case param.Vol:
		if param.VolImageModels[model] {
			return model
		}
		return param.DoubaoSeed16VisionPro
	case param.ChatAnyWhere:
		return model
	}
	
	return ""
}

func GetUsingVideoModel(ty string, model string) string {
	switch ty {
	case param.Gemini:
		if param.GeminiVideoModels[model] {
			return model
		}
		return param.GeminiVideoVeo2
	
	case param.Aliyun:
		if param.AliyunVideoModels[model] {
			return model
		}
		return param.Wan2_5T2VPreview
	case param.AI302:
		return model
	case param.Vol:
		if param.VolVideoModels[model] {
			return model
		}
		return param.DoubaoSeedance1_0Pro
	}
	
	return ""
}

func GetUsingRecModel(ty string, model string) string {
	switch ty {
	case param.Gemini:
		if param.GeminiRecModels[model] {
			return model
		}
		return param.ModelGemini20Flash
	
	case param.OpenAi:
		if param.OpenAiRecModels[model] {
			return model
		}
		return param.ChatGPT4_0
	case param.Aliyun:
		if param.AliyunRecModels[model] {
			return model
		}
		return param.QwenVlMax
	case param.AI302:
		return model
	case param.Vol:
		if param.VolRecModels[model] {
			return model
		}
		return param.DoubaoSeed16VisionPro
	}
	
	return ""
}

func GetUsingTxtModel(ty string, model string) string {
	switch ty {
	case param.DeepSeek:
		if param.DeepseekModels[model] {
			return model
		}
		return godeepseek.DeepSeekChat
	case param.Gemini:
		if param.GeminiModels[model] {
			return model
		}
		return param.ModelGemini20Flash
	case param.Vol:
		if param.VolModels[model] {
			return model
		}
		return param.ModelDeepSeekR1_528
	case param.Aliyun:
		if param.AliyunModel[model] {
			return model
		}
		return qwen.QwenMax
	}
	
	return model
}
