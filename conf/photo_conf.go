package conf

import (
	"flag"
	"os"
	"strconv"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

type PhotoConf struct {
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
}

var PhotoConfInfo = new(PhotoConf)

func InitPhotoConf() {
	PhotoConfInfo.ReqKey = flag.String("req_key", "high_aes_general_v21_L", "request key")
	PhotoConfInfo.ModelVersion = flag.String("model_version", "general_v2.1_L", "model version")
	PhotoConfInfo.ReqScheduleConf = flag.String("req_schedule_conf", "general_v20_9B_pe", "request schedule conf")
	PhotoConfInfo.Seed = flag.Int("seed", -1, "seed for random seed")
	PhotoConfInfo.Scale = flag.Float64("scale", 3.5, "scale factor")
	PhotoConfInfo.DDIMSteps = flag.Int("ddim_steps", 25, "ddim steps")
	PhotoConfInfo.Width = flag.Int("width", 512, "width of the image")
	PhotoConfInfo.Height = flag.Int("height", 512, "height of the image")
	PhotoConfInfo.UsePreLLM = flag.Bool("use_pre_llm", true, "use pre llm")
	PhotoConfInfo.UseSr = flag.Bool("use_sr", true, "use super resolution")
	PhotoConfInfo.ReturnUrl = flag.Bool("return_url", true, "return url")
	PhotoConfInfo.AddLogo = flag.Bool("add_logo", false, "add logo")
	PhotoConfInfo.Position = flag.String("position", "", "position")
	PhotoConfInfo.Language = flag.Int("language", 1, "language")
	PhotoConfInfo.Opacity = flag.Float64("opacity", 0.3, "opacity")
	PhotoConfInfo.LogoTextContent = flag.String("logo_text_content", "", "logo text content")
}

func EnvPhotoConf() {
	if os.Getenv("REQ_KEY") != "" {
		*PhotoConfInfo.ReqKey = os.Getenv("REQ_KEY")
	}
	
	if os.Getenv("MODEL_VERSION") != "" {
		*PhotoConfInfo.ModelVersion = os.Getenv("MODEL_VERSION")
	}
	
	if os.Getenv("REQ_SCHEDULE_CONF") != "" {
		*PhotoConfInfo.ReqScheduleConf = os.Getenv("REQ_SCHEDULE_CONF")
	}
	
	if os.Getenv("SEED") != "" {
		*PhotoConfInfo.Seed, _ = strconv.Atoi(os.Getenv("SEED"))
	}
	
	if os.Getenv("SCALE") != "" {
		*PhotoConfInfo.Scale, _ = strconv.ParseFloat(os.Getenv("SCALE"), 64)
	}
	
	if os.Getenv("DDIM_Steps") != "" {
		*PhotoConfInfo.DDIMSteps, _ = strconv.Atoi(os.Getenv("DDIM_Steps"))
	}
	
	if os.Getenv("WIDTH") != "" {
		*PhotoConfInfo.Width, _ = strconv.Atoi(os.Getenv("WIDTH"))
	}
	
	if os.Getenv("Height") != "" {
		*PhotoConfInfo.Height, _ = strconv.Atoi(os.Getenv("Height"))
	}
	
	if os.Getenv("UsePreLLM") != "" {
		*PhotoConfInfo.UsePreLLM, _ = strconv.ParseBool(os.Getenv("UsePreLLM"))
	}
	
	if os.Getenv("UseSr") != "" {
		*PhotoConfInfo.UseSr, _ = strconv.ParseBool(os.Getenv("UseSr"))
	}
	
	if os.Getenv("ReturnUrl") != "" {
		*PhotoConfInfo.ReturnUrl, _ = strconv.ParseBool(os.Getenv("ReturnUrl"))
	}
	
	if os.Getenv("AddLogo") != "" {
		*PhotoConfInfo.AddLogo, _ = strconv.ParseBool(os.Getenv("AddLogo"))
	}
	
	if os.Getenv("Position") != "" {
		*PhotoConfInfo.Position = os.Getenv("Position")
	}
	
	if os.Getenv("Language") != "" {
		*PhotoConfInfo.Language, _ = strconv.Atoi(os.Getenv("Language"))
	}
	
	if os.Getenv("Opacity") != "" {
		*PhotoConfInfo.Opacity, _ = strconv.ParseFloat(os.Getenv("Opacity"), 64)
	}
	
	if os.Getenv("LogoTextContent") != "" {
		*PhotoConfInfo.LogoTextContent = os.Getenv("LogoTextContent")
	}
	
	logger.Info("PHOTO_CONF", "ReqKey", *PhotoConfInfo.ReqKey)
	logger.Info("PHOTO_CONF", "ModelVersion", *PhotoConfInfo.ModelVersion)
	logger.Info("PHOTO_CONF", "ReqScheduleConf", *PhotoConfInfo.ReqScheduleConf)
	logger.Info("PHOTO_CONF", "Seed", *PhotoConfInfo.Seed)
	logger.Info("PHOTO_CONF", "Width", *PhotoConfInfo.Width)
	logger.Info("PHOTO_CONF", "Height", *PhotoConfInfo.Height)
	logger.Info("PHOTO_CONF", "Scale", *PhotoConfInfo.Scale)
	logger.Info("PHOTO_CONF", "DDIMSteps", *PhotoConfInfo.DDIMSteps)
	logger.Info("PHOTO_CONF", "UsePreLLM", *PhotoConfInfo.UsePreLLM)
	logger.Info("PHOTO_CONF", "UseSr", *PhotoConfInfo.UseSr)
	logger.Info("PHOTO_CONF", "ReturnUrl", *PhotoConfInfo.ReturnUrl)
	logger.Info("PHOTO_CONF", "AddLogo", *PhotoConfInfo.AddLogo)
	logger.Info("PHOTO_CONF", "Position", *PhotoConfInfo.Position)
	logger.Info("PHOTO_CONF", "Language", *PhotoConfInfo.Language)
	logger.Info("PHOTO_CONF", "Opacity", *PhotoConfInfo.Opacity)
	logger.Info("PHOTO_CONF", "LogoTextContent", *PhotoConfInfo.LogoTextContent)
}
