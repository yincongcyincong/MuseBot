package conf

import (
	"flag"
	"os"
	"strconv"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

type VideoConf struct {
	VideoModel *string
	Radio      *string
	Duration   *int
	FPS        *int
	Resolution *string
	Watermark  *bool
}

var (
	VideoConfInfo = new(VideoConf)
)

func InitVideoConf() {
	VideoConfInfo.VideoModel = flag.String("video_model", "doubao-seaweed-241128", "video model")
	VideoConfInfo.Radio = flag.String("radio", "1:1", "the width to height ratio")
	VideoConfInfo.Duration = flag.Int("duration", 5, "the duration in seconds, only support 5s / 10s")
	VideoConfInfo.FPS = flag.Int("fps", 24, "the frame per second")
	VideoConfInfo.Resolution = flag.String("resolution", "480p", "the resolution of video, only support 480p / 720p")
	VideoConfInfo.Watermark = flag.Bool("watermark", false, "include watermark")
	
}

func EnvVideoConf() {
	if os.Getenv("VIDEO_MODEL") != "" {
		*VideoConfInfo.VideoModel = os.Getenv("VIDEO_MODEL")
	}
	
	if os.Getenv("RADIO") != "" {
		*VideoConfInfo.Radio = os.Getenv("RADIO")
	}
	
	if os.Getenv("DURATION") != "" {
		*VideoConfInfo.Duration, _ = strconv.Atoi(os.Getenv("DURATION"))
	}
	
	if os.Getenv("FPS") != "" {
		*VideoConfInfo.FPS, _ = strconv.Atoi(os.Getenv("FPS"))
	}
	
	if os.Getenv("RESOLUTION") != "" {
		*VideoConfInfo.Resolution = os.Getenv("RESOLUTION")
	}
	
	if os.Getenv("WATERMARK") != "" {
		*VideoConfInfo.Watermark, _ = strconv.ParseBool(os.Getenv("WATERMARK"))
	}
	
	logger.Info("VIDEO_CONF", "VIDEO_MODEL", *VideoConfInfo.VideoModel)
	logger.Info("VIDEO_CONF", "RADIO", *VideoConfInfo.Radio)
	logger.Info("VIDEO_CONF", "DURATION", *VideoConfInfo.Duration)
	logger.Info("VIDEO_CONF", "FPS", *VideoConfInfo.FPS)
	logger.Info("VIDEO_CONF", "RESOLUTION", *VideoConfInfo.Resolution)
	logger.Info("VIDEO_CONF", "WATERMARK", *VideoConfInfo.Watermark)
}
