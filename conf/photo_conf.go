package conf

import (
	"flag"
	"os"
	"strconv"

	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

var (
	ReqKey          *string
	ModelVersion    *string
	ReqScheduleConf *string
	Seed            *int
	Scale           *float64
	DDIMSteps       *int
	Width           *int
	Height          *int
	UsePreLLM       *bool
	UseSr           *bool
	ReturnUrl       *bool
	AddLogo         *bool
	Position        *string
	Language        *int
	Opacity         *float64
	LogoTextContent *string
)

func InitPhotoConf() {
	ReqKey = flag.String("req_key", "high_aes_general_v21_L", "request key")
	ModelVersion = flag.String("model_version", "general_v2.1_L", "model version")
	ReqScheduleConf = flag.String("req_schedule_conf", "general_v20_9B_pe", "request schedule conf")
	Seed = flag.Int("seed", -1, "seed for random seed")
	Scale = flag.Float64("scale", 3.5, "scale factor")
	DDIMSteps = flag.Int("ddim_steps", 25, "ddim steps")
	Width = flag.Int("width", 512, "width of the image")
	Height = flag.Int("height", 512, "height of the image")
	UsePreLLM = flag.Bool("use_pre_llm", true, "use pre llm")
	UseSr = flag.Bool("use_sr", true, "use super resolution")
	ReturnUrl = flag.Bool("return_url", true, "return url")
	AddLogo = flag.Bool("add_logo", false, "add logo")
	Position = flag.String("position", "", "position")
	Language = flag.Int("language", 1, "language")
	Opacity = flag.Float64("opacity", 0.3, "opacity")
	LogoTextContent = flag.String("logo_text_content", "", "logo text content")

	if os.Getenv("REQ_KEY") != "" {
		*ReqKey = os.Getenv("REQ_KEY")
	}

	if os.Getenv("MODEL_VERSION") != "" {
		*ModelVersion = os.Getenv("MODEL_VERSION")
	}

	if os.Getenv("REQ_SCHEDULE_CONF") != "" {
		*ReqScheduleConf = os.Getenv("REQ_SCHEDULE_CONF")
	}

	if os.Getenv("SEED") != "" {
		*Seed, _ = strconv.Atoi(os.Getenv("SEED"))
	}

	if os.Getenv("SCALE") != "" {
		*Scale, _ = strconv.ParseFloat(os.Getenv("SCALE"), 64)
	}

	if os.Getenv("DDIM_Steps") != "" {
		*DDIMSteps, _ = strconv.Atoi(os.Getenv("DDIM_Steps"))
	}

	if os.Getenv("WIDTH") != "" {
		*Width, _ = strconv.Atoi(os.Getenv("WIDTH"))
	}

	if os.Getenv("Height") != "" {
		*Height, _ = strconv.Atoi(os.Getenv("Height"))
	}

	if os.Getenv("UsePreLLM") != "" {
		*UsePreLLM, _ = strconv.ParseBool(os.Getenv("UsePreLLM"))
	}

	if os.Getenv("UseSr") != "" {
		*UseSr, _ = strconv.ParseBool(os.Getenv("UseSr"))
	}

	if os.Getenv("ReturnUrl") != "" {
		*ReturnUrl, _ = strconv.ParseBool(os.Getenv("ReturnUrl"))
	}

	if os.Getenv("AddLogo") != "" {
		*AddLogo, _ = strconv.ParseBool(os.Getenv("AddLogo"))
	}

	if os.Getenv("Position") != "" {
		*Position = os.Getenv("Position")
	}

	if os.Getenv("Language") != "" {
		*Language, _ = strconv.Atoi(os.Getenv("Language"))
	}

	if os.Getenv("Opacity") != "" {
		*Opacity, _ = strconv.ParseFloat(os.Getenv("Opacity"), 64)
	}

	if os.Getenv("LogoTextContent") != "" {
		*LogoTextContent = os.Getenv("LogoTextContent")
	}

	logger.Info("PHOTO_CONF", "ReqKey", *ReqKey)
	logger.Info("PHOTO_CONF", "ModelVersion", *ModelVersion)
	logger.Info("PHOTO_CONF", "ReqScheduleConf", *ReqScheduleConf)
	logger.Info("PHOTO_CONF", "Seed", *Seed)
	logger.Info("PHOTO_CONF", "Width", *Width)
	logger.Info("PHOTO_CONF", "Height", *Height)
	logger.Info("PHOTO_CONF", "Scale", *Scale)
	logger.Info("PHOTO_CONF", "DDIMSteps", *DDIMSteps)
	logger.Info("PHOTO_CONF", "UsePreLLM", *UsePreLLM)
	logger.Info("PHOTO_CONF", "UseSr", *UseSr)
	logger.Info("PHOTO_CONF", "ReturnUrl", *ReturnUrl)
	logger.Info("PHOTO_CONF", "AddLogo", *AddLogo)
	logger.Info("PHOTO_CONF", "Position", *Position)
	logger.Info("PHOTO_CONF", "Language", *Language)
	logger.Info("PHOTO_CONF", "Opacity", *Opacity)
	logger.Info("PHOTO_CONF", "LogoTextContent", *LogoTextContent)

}
