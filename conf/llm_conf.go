package conf

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

type LLMConf struct {
	FrequencyPenalty *float64 `json:"frequency_penalty"`
	MaxTokens        *int     `json:"max_tokens"`
	PresencePenalty  *float64 `json:"presence_penalty"`
	Temperature      *float64 `json:"temperature"`
	TopP             *float64 `json:"top_p"`
	Stop             []string `json:"stop"`
	LogProbs         *bool    `json:"log_probs"`
	TopLogProbs      *int     `json:"top_log_probs"`
	
	stop *string
}

var (
	LLMConfInfo = new(LLMConf)
)

func InitLLMConf() {
	
	LLMConfInfo.FrequencyPenalty = flag.Float64("frequency_penalty", 0.0, "frequency penalty")
	LLMConfInfo.MaxTokens = flag.Int("max_tokens", 2048, "maximum number of tokens")
	LLMConfInfo.PresencePenalty = flag.Float64("presence_penalty", 0.0, "presence penalty")
	LLMConfInfo.Temperature = flag.Float64("temperature", 0.7, "temperature")
	LLMConfInfo.TopP = flag.Float64("top_p", 0.9, "top p")
	LLMConfInfo.LogProbs = flag.Bool("log_probs", false, "log probs")
	LLMConfInfo.TopLogProbs = flag.Int("top_log_probs", 0, "number of top log probs to return")
	
	LLMConfInfo.stop = flag.String("stop", "", "stop sequence")
}

func EnvLLMConf() {
	if os.Getenv("FREQUENCY_PENALTY") != "" {
		*LLMConfInfo.FrequencyPenalty, _ = strconv.ParseFloat(os.Getenv("FREQUENCY_PENALTY"), 64)
	}
	
	if os.Getenv("MAX_TOKENS") != "" {
		*LLMConfInfo.MaxTokens, _ = strconv.Atoi(os.Getenv("MAX_TOKENS"))
	}
	
	if os.Getenv("PRESENCE_PENALTY") != "" {
		*LLMConfInfo.PresencePenalty, _ = strconv.ParseFloat(os.Getenv("PRESENCE_PENALTY"), 64)
	}
	
	if os.Getenv("TEMPERATURE") != "" {
		*LLMConfInfo.Temperature, _ = strconv.ParseFloat(os.Getenv("TEMPERATURE"), 64)
	}
	
	if os.Getenv("TOP_P") != "" {
		*LLMConfInfo.TopP, _ = strconv.ParseFloat(os.Getenv("TOP_P"), 64)
	}
	
	if os.Getenv("STOP") != "" {
		*LLMConfInfo.stop = os.Getenv("STOP")
	}
	
	if os.Getenv("LOG_PROBS") != "" {
		*LLMConfInfo.LogProbs, _ = strconv.ParseBool(os.Getenv("LOG_PROBS"))
	}
	
	if os.Getenv("TOP_LOG_PROBS") != "" {
		*LLMConfInfo.TopLogProbs, _ = strconv.Atoi(os.Getenv("TOP_LOG_PROBS"))
	}
	
	for _, s := range strings.Split(*LLMConfInfo.stop, ",") {
		if s != "" {
			LLMConfInfo.Stop = append(LLMConfInfo.Stop, s)
		}
		
	}
	
}
