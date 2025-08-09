package conf

import (
	"flag"
	"os"
	"strconv"
	
	"github.com/yincongcyincong/MuseBot/logger"
)

type VideoConf struct {
	VideoModel *string `json:"video_model"`
	Radio      *string `json:"radio"`
	Duration   *int    `json:"duration"`
	FPS        *int    `json:"fps"`
	Resolution *string `json:"resolution"`
	Watermark  *bool   `json:"watermark"`
	
	GeminiVideoModel            *string `json:"gemini_video_model"`
	GeminiVideoAspectRatio      *string `json:"gemini_video_aspect_ratio"`
	GeminiVideoDurationSeconds  int32   `json:"gemini_video_duration_seconds"`
	GeminiVideoPersonGeneration *string `json:"gemini_video_person_generation"`
}

var (
	VideoConfInfo = new(VideoConf)
)

func InitVideoConf() {
	VideoConfInfo.VideoModel = flag.String("video_model", "doubao-seedance-1-0-pro-250528", "video model")
	VideoConfInfo.Radio = flag.String("radio", "1:1", "the width to height ratio")
	VideoConfInfo.Duration = flag.Int("duration", 5, "the duration in seconds, only support 5s / 10s")
	VideoConfInfo.FPS = flag.Int("fps", 24, "the frame per second")
	VideoConfInfo.Resolution = flag.String("resolution", "480p", "the resolution of video, only support 480p / 720p")
	VideoConfInfo.Watermark = flag.Bool("watermark", false, "include watermark")
	
	VideoConfInfo.GeminiVideoModel = flag.String("gemini_video_model", "veo-2.0-generate-001", "create video model")
	VideoConfInfo.GeminiVideoAspectRatio = flag.String("gemini_video_aspect_ratio", "16:9", "gemini video ratio")
	VideoConfInfo.GeminiVideoDurationSeconds = int32(*flag.Int("gemini_video_duration_seconds", 0, "gemini video duration"))
	VideoConfInfo.GeminiVideoPersonGeneration = flag.String("gemini_video_person_generation", "allow_all", "gemini video can generate person or not: allow_all or ")
	
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
	
	if os.Getenv("GEMINI_VIDEO_MODEL") != "" {
		*VideoConfInfo.GeminiVideoModel = os.Getenv("GEMINI_VIDEO_MODEL")
	}
	if os.Getenv("GEMINI_VIDEO_PERSON_GENERATION") != "" {
		*VideoConfInfo.GeminiVideoPersonGeneration = os.Getenv("GEMINI_VIDEO_PERSON_GENERATION")
	}
	if os.Getenv("GEMINI_VIDEO_ASPECT_RATIO") != "" {
		*VideoConfInfo.GeminiVideoAspectRatio = os.Getenv("GEMINI_VIDEO_ASPECT_RATIO")
	}
	if os.Getenv("GEMINI_VIDEO_DURATION_SECONDS") != "" {
		tmp, _ := strconv.Atoi(os.Getenv("GEMINI_VIDEO_DURATION_SECONDS"))
		VideoConfInfo.GeminiVideoDurationSeconds = int32(tmp)
	}
	
	logger.Info("VIDEO_CONF", "VIDEO_MODEL", *VideoConfInfo.VideoModel)
	logger.Info("VIDEO_CONF", "RADIO", *VideoConfInfo.Radio)
	logger.Info("VIDEO_CONF", "DURATION", *VideoConfInfo.Duration)
	logger.Info("VIDEO_CONF", "FPS", *VideoConfInfo.FPS)
	logger.Info("VIDEO_CONF", "RESOLUTION", *VideoConfInfo.Resolution)
	logger.Info("VIDEO_CONF", "WATERMARK", *VideoConfInfo.Watermark)
	logger.Info("AUDIO_CONF", "GeminiVideoModel", *VideoConfInfo.GeminiVideoModel)
	logger.Info("AUDIO_CONF", "GeminiVideoPersonGeneration", *VideoConfInfo.GeminiVideoPersonGeneration)
	logger.Info("AUDIO_CONF", "GeminiVideoAspectRatio", *VideoConfInfo.GeminiVideoAspectRatio)
	logger.Info("AUDIO_CONF", "GeminiVideoDurationSeconds", VideoConfInfo.GeminiVideoDurationSeconds)
}
