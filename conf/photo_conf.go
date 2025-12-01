package conf

import (
	"flag"
	"os"
	"strconv"
	
	"github.com/sashabaranov/go-openai"
)

type PhotoConf struct {
	ReqKey          string  `json:"req_key"`
	ModelVersion    string  `json:"model_version"`
	ReqScheduleConf string  `json:"req_schedule_conf"`
	Seed            int     `json:"seed"`
	Scale           float64 `json:"scale"`
	DDIMSteps       int     `json:"ddim_steps"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	UsePreLLM       bool    `json:"use_pre_llm"`
	UseSr           bool    `json:"use_sr"`
	ReturnUrl       bool    `json:"return_url"`
	AddLogo         bool    `json:"add_logo"`
	Position        string  `json:"position"`
	Language        int     `json:"language"`
	Opacity         float64 `json:"opacity"`
	LogoTextContent string  `json:"logo_text_content"`
	
	VolImageModel string `json:"vol_image_model"`
	VolRecModel   string `json:"vol_rec_model"`
	
	GeminiImageModel string `json:"gemini_image_model"`
	GeminiRecModel   string `json:"gemini_rec_model"`
	
	OpenAIImageModel string `json:"openai_image_model"`
	OpenAIRecModel   string `json:"openai_rec_model"`
	OpenAIImageSize  string `json:"openai_image_size"`
	OpenAIImageStyle string `json:"openai_image_style"`
	
	MixImageModel string `json:"mix_image_model"`
	MixRecModel   string `json:"mix_rec_model"`
	
	AliyunImageModel string `json:"aliyun_image_model"`
	AliyunRecModel   string `json:"aliyun_rec_model"`
}

var PhotoConfInfo = new(PhotoConf)

func InitPhotoConf() {
	flag.StringVar(&PhotoConfInfo.ReqKey, "req_key", "high_aes_general_v21_L", "request key")
	flag.StringVar(&PhotoConfInfo.ModelVersion, "model_version", "general_v2.1_L", "model version")
	flag.StringVar(&PhotoConfInfo.ReqScheduleConf, "req_schedule_conf", "general_v20_9B_pe", "request schedule conf")
	
	flag.IntVar(&PhotoConfInfo.Seed, "seed", -1, "seed for random seed")
	flag.Float64Var(&PhotoConfInfo.Scale, "scale", 3.5, "scale factor")
	flag.IntVar(&PhotoConfInfo.DDIMSteps, "ddim_steps", 25, "ddim steps")
	flag.IntVar(&PhotoConfInfo.Width, "width", 512, "width of the image")
	flag.IntVar(&PhotoConfInfo.Height, "height", 512, "height of the image")
	
	flag.BoolVar(&PhotoConfInfo.UsePreLLM, "use_pre_llm", true, "use pre llm")
	flag.BoolVar(&PhotoConfInfo.UseSr, "use_sr", true, "use super resolution")
	flag.BoolVar(&PhotoConfInfo.ReturnUrl, "return_url", true, "return url")
	flag.BoolVar(&PhotoConfInfo.AddLogo, "add_logo", false, "add logo")
	
	flag.StringVar(&PhotoConfInfo.Position, "position", "", "position")
	flag.IntVar(&PhotoConfInfo.Language, "language", 1, "language")
	flag.Float64Var(&PhotoConfInfo.Opacity, "opacity", 0.3, "opacity")
	flag.StringVar(&PhotoConfInfo.LogoTextContent, "logo_text_content", "", "logo text content")
	
	flag.StringVar(&PhotoConfInfo.GeminiImageModel, "gemini_image_model", "gemini-2.0-flash-preview-image-generation", "gemini create photo model")
	flag.StringVar(&PhotoConfInfo.GeminiRecModel, "gemini_rec_model", "gemini-2.0-flash", "gemini recognize photo model")
	
	flag.StringVar(&PhotoConfInfo.OpenAIRecModel, "openai_rec_model", "chatgpt-4o-latest", "openai create photo model")
	flag.StringVar(&PhotoConfInfo.OpenAIImageModel, "openai_image_model", "gpt-image-1", "openai create photo model")
	flag.StringVar(&PhotoConfInfo.OpenAIImageSize, "openai_image_size", openai.CreateImageSize1024x1024, "openai image size")
	flag.StringVar(&PhotoConfInfo.OpenAIImageStyle, "openai_image_style", "", "openai image style")
	
	flag.StringVar(&PhotoConfInfo.VolImageModel, "vol_image_model", "doubao-seed-1-6-250615", "vol image model")
	flag.StringVar(&PhotoConfInfo.VolRecModel, "vol_rec_model", "doubao-seed-1-6-250615", "vol recognize photo model")
	
	flag.StringVar(&PhotoConfInfo.MixImageModel, "mix_image_model", "gpt-image-1", "ai302/openrouter image model")
	flag.StringVar(&PhotoConfInfo.MixRecModel, "mix_rec_model", "chatgpt-4o-latest", "ai302/openrouter recognize photo model")
	
	flag.StringVar(&PhotoConfInfo.AliyunImageModel, "aliyun_image_model", "qwen-image-plus", "aliyun image model")
	flag.StringVar(&PhotoConfInfo.AliyunRecModel, "aliyun_rec_model", "qwen-vl-max-latest", "aliyun recognize photo model")
	
}

func EnvPhotoConf() {
	if os.Getenv("REQ_KEY") != "" {
		PhotoConfInfo.ReqKey = os.Getenv("REQ_KEY")
	}
	
	if os.Getenv("MODEL_VERSION") != "" {
		PhotoConfInfo.ModelVersion = os.Getenv("MODEL_VERSION")
	}
	
	if os.Getenv("REQ_SCHEDULE_CONF") != "" {
		PhotoConfInfo.ReqScheduleConf = os.Getenv("REQ_SCHEDULE_CONF")
	}
	
	if os.Getenv("SEED") != "" {
		PhotoConfInfo.Seed, _ = strconv.Atoi(os.Getenv("SEED"))
	}
	
	if os.Getenv("SCALE") != "" {
		PhotoConfInfo.Scale, _ = strconv.ParseFloat(os.Getenv("SCALE"), 64)
	}
	
	if os.Getenv("DDIM_Steps") != "" {
		PhotoConfInfo.DDIMSteps, _ = strconv.Atoi(os.Getenv("DDIM_Steps"))
	}
	
	if os.Getenv("WIDTH") != "" {
		PhotoConfInfo.Width, _ = strconv.Atoi(os.Getenv("WIDTH"))
	}
	
	if os.Getenv("HEIGHT") != "" {
		PhotoConfInfo.Height, _ = strconv.Atoi(os.Getenv("HEIGHT"))
	}
	
	if os.Getenv("USE_PER_LLM") != "" {
		PhotoConfInfo.UsePreLLM, _ = strconv.ParseBool(os.Getenv("USE_PER_LLM"))
	}
	
	if os.Getenv("USE_SR") != "" {
		PhotoConfInfo.UseSr, _ = strconv.ParseBool(os.Getenv("USE_SR"))
	}
	
	if os.Getenv("RETURN_URL") != "" {
		PhotoConfInfo.ReturnUrl, _ = strconv.ParseBool(os.Getenv("RETURN_URL"))
	}
	
	if os.Getenv("ADD_LOGO") != "" {
		PhotoConfInfo.AddLogo, _ = strconv.ParseBool(os.Getenv("ADD_LOGO"))
	}
	
	if os.Getenv("POSITION") != "" {
		PhotoConfInfo.Position = os.Getenv("POSITION")
	}
	
	if os.Getenv("PHOTO_LANGUAGE") != "" {
		PhotoConfInfo.Language, _ = strconv.Atoi(os.Getenv("PHOTO_LANGUAGE"))
	}
	
	if os.Getenv("OPACITY") != "" {
		PhotoConfInfo.Opacity, _ = strconv.ParseFloat(os.Getenv("OPACITY"), 64)
	}
	
	if os.Getenv("LOGO_TEXT_CONTENT") != "" {
		PhotoConfInfo.LogoTextContent = os.Getenv("LOGO_TEXT_CONTENT")
	}
	
	if os.Getenv("GEMINI_IMAGE_MODEL") != "" {
		PhotoConfInfo.GeminiImageModel = os.Getenv("GEMINI_IMAGE_MODEL")
	}
	
	if os.Getenv("GEMINI_REC_MODEL") != "" {
		PhotoConfInfo.GeminiRecModel = os.Getenv("GEMINI_REC_MODEL")
	}
	
	if os.Getenv("OPENAI_REC_MODEL") != "" {
		PhotoConfInfo.OpenAIRecModel = os.Getenv("OPENAI_REC_MODEL")
	}
	
	if os.Getenv("OPENAI_IMAGE_MODEL") != "" {
		PhotoConfInfo.OpenAIImageModel = os.Getenv("OPENAI_IMAGE_MODEL")
	}
	
	if os.Getenv("OPENAI_IMAGE_SIZE") != "" {
		PhotoConfInfo.OpenAIImageSize = os.Getenv("OPENAI_IMAGE_SIZE")
	}
	
	if os.Getenv("OPENAI_IMAGE_STYLE") != "" {
		PhotoConfInfo.OpenAIImageStyle = os.Getenv("OPENAI_IMAGE_STYLE")
	}
	
	if os.Getenv("VOL_IMAGE_MODEL") != "" {
		PhotoConfInfo.VolImageModel = os.Getenv("VOL_IMAGE_MODEL")
	}
	
	if os.Getenv("VOL_REC_MODEL") != "" {
		PhotoConfInfo.VolRecModel = os.Getenv("VOL_REC_MODEL")
	}
	
	if os.Getenv("MIX_IMAGE_MODEL") != "" {
		PhotoConfInfo.MixImageModel = os.Getenv("MIX_IMAGE_MODEL")
	}
	
	if os.Getenv("MIX_REC_MODEL") != "" {
		PhotoConfInfo.MixRecModel = os.Getenv("MIX_REC_MODEL")
	}
	
	if os.Getenv("ALIYUN_IMAGE_MODEL") != "" {
		PhotoConfInfo.AliyunImageModel = os.Getenv("ALIYUN_IMAGE_MODEL")
	}
	
	if os.Getenv("ALIYUN_REC_MODEL") != "" {
		PhotoConfInfo.AliyunRecModel = os.Getenv("ALIYUN_REC_MODEL")
	}
}
