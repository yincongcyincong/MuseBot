package conf

import (
	"flag"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"os"
	"strconv"
)

var (
	VideoModel *string
	Radio      *string
	Duration   *int
	FPS        *int
	Resolution *string
	Watermark  *bool
)

func InitVideoConf() {
	VideoModel = flag.String("video_model", "doubao-seaweed-241128", "video model")
	Radio = flag.String("radio", "1:1", "the width to height ratio")
	Duration = flag.Int("duration", 5, "the duration in seconds, only support 5s / 10s")
	FPS = flag.Int("fps", 24, "the frame per second")
	Resolution = flag.String("resolution", "480p", "the resolution of video, only support 480p / 720p")
	Watermark = flag.Bool("watermark", false, "include watermark")

	if os.Getenv("VIDEO_MODEL") != "" {
		*VideoModel = os.Getenv("VIDEO_MODEL")
	}

	if os.Getenv("RADIO") != "" {
		*Radio = os.Getenv("RADIO")
	}

	if os.Getenv("DURATION") != "" {
		*Duration, _ = strconv.Atoi(os.Getenv("DURATION"))
	}

	if os.Getenv("FPS") != "" {
		*FPS, _ = strconv.Atoi(os.Getenv("FPS"))
	}

	if os.Getenv("RESOLUTION") != "" {
		*Resolution = os.Getenv("RESOLUTION")
	}

	if os.Getenv("WATERMARK") != "" {
		*Watermark, _ = strconv.ParseBool(os.Getenv("WATERMARK"))
	}

	logger.Info("VIDEO_CONF", "VIDEO_MODEL", *VideoModel)
	logger.Info("VIDEO_CONF", "RADIO", *Radio)
	logger.Info("VIDEO_CONF", "DURATION", *Duration)
	logger.Info("VIDEO_CONF", "FPS", *FPS)
	logger.Info("VIDEO_CONF", "RESOLUTION", *Resolution)
	logger.Info("VIDEO_CONF", "WATERMARK", *Watermark)

}
