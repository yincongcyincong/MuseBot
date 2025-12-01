package conf

import (
	"flag"
	"os"
	"strconv"
)

type VideoConf struct {
	VolVideoModel    string `json:"vol_video_model"`
	GeminiVideoModel string `json:"gemini_video_model"`
	AI302VideoModel  string `json:"ai_302_video_model"`
	AliyunVideoModel string `json:"aliyun_video_model"`
	
	Radio      string `json:"radio"`
	Duration   int    `json:"duration"`
	FPS        int    `json:"fps"`
	Resolution string `json:"resolution"`
	Watermark  bool   `json:"watermark"`
}

var (
	VideoConfInfo = new(VideoConf)
)

func InitVideoConf() {
	flag.StringVar(&VideoConfInfo.VolVideoModel, "vol_video_model", "doubao-seedance-1-0-pro-250528", "video model")
	flag.StringVar(&VideoConfInfo.Radio, "radio", "1:1", "the width to height ratio")
	flag.IntVar(&VideoConfInfo.Duration, "duration", 5, "the duration in seconds, only support 5s / 10s")
	flag.IntVar(&VideoConfInfo.FPS, "fps", 24, "the frame per second")
	flag.StringVar(&VideoConfInfo.Resolution, "resolution", "480p", "the resolution of video, only support 480p / 720p")
	flag.BoolVar(&VideoConfInfo.Watermark, "watermark", false, "include watermark")
	
	flag.StringVar(&VideoConfInfo.GeminiVideoModel, "gemini_video_model", "veo-2.0-generate-001", "create video model")
	
	flag.StringVar(&VideoConfInfo.AI302VideoModel, "ai_302_video_model", "luma_video", "create video model")
	
	flag.StringVar(&VideoConfInfo.AliyunVideoModel, "aliyun_video_model", "wan2.5-t2v-preview", "create video model")
}

func EnvVideoConf() {
	if os.Getenv("VOL_VIDEO_MODEL") != "" {
		VideoConfInfo.VolVideoModel = os.Getenv("VOL_VIDEO_MODEL")
	}
	
	if os.Getenv("RADIO") != "" {
		VideoConfInfo.Radio = os.Getenv("RADIO")
	}
	
	if os.Getenv("DURATION") != "" {
		VideoConfInfo.Duration, _ = strconv.Atoi(os.Getenv("DURATION"))
	}
	
	if os.Getenv("FPS") != "" {
		VideoConfInfo.FPS, _ = strconv.Atoi(os.Getenv("FPS"))
	}
	
	if os.Getenv("RESOLUTION") != "" {
		VideoConfInfo.Resolution = os.Getenv("RESOLUTION")
	}
	
	if os.Getenv("WATERMARK") != "" {
		VideoConfInfo.Watermark, _ = strconv.ParseBool(os.Getenv("WATERMARK"))
	}
	
	if os.Getenv("GEMINI_VIDEO_MODEL") != "" {
		VideoConfInfo.GeminiVideoModel = os.Getenv("GEMINI_VIDEO_MODEL")
	}
	
	if os.Getenv("302_AI_VIDEO_MODEL") != "" {
		VideoConfInfo.AI302VideoModel = os.Getenv("302_AI_VIDEO_MODEL")
	}
}
