package conf

import (
	"flag"
	"os"
	"strconv"
	"strings"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
)

type LLMConf struct {
	FrequencyPenalty *float64
	MaxTokens        *int
	PresencePenalty  *float64
	Temperature      *float64
	TopP             *float64
	Stop             []string
	LogProbs         *bool
	TopLogProbs      *int
	
	stop *string
}

var (
	LLMConfInfo = new(LLMConf)
)

func InitDeepseekConf() {
	
	LLMConfInfo.FrequencyPenalty = flag.Float64("frequency_penalty", 0.0, "frequency penalty")
	LLMConfInfo.MaxTokens = flag.Int("max_tokens", 2048, "maximum number of tokens")
	LLMConfInfo.PresencePenalty = flag.Float64("presence_penalty", 0.0, "presence penalty")
	LLMConfInfo.Temperature = flag.Float64("temperature", 0.7, "temperature")
	LLMConfInfo.TopP = flag.Float64("top_p", 0.9, "top p")
	LLMConfInfo.LogProbs = flag.Bool("log_probs", false, "log probs")
	LLMConfInfo.TopLogProbs = flag.Int("top_log_probs", 0, "number of top log probs to return")
	
	LLMConfInfo.stop = flag.String("stop", "", "stop sequence")
}

func EnvDeepseekConf() {
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
	
	logger.Info("LLM_CONF", "FrequencyPenalty", *LLMConfInfo.FrequencyPenalty)
	logger.Info("LLM_CONF", "MaxTokens", *LLMConfInfo.MaxTokens)
	logger.Info("LLM_CONF", "PresencePenalty", *LLMConfInfo.PresencePenalty)
	logger.Info("LLM_CONF", "Temperature", *LLMConfInfo.Temperature)
	logger.Info("LLM_CONF", "TopP", *LLMConfInfo.TopP)
	logger.Info("LLM_CONF", "Stop", *LLMConfInfo.stop)
	logger.Info("LLM_CONF", "LogProbs", *LLMConfInfo.LogProbs)
	logger.Info("LLM_CONF", "TopLogProbs", *LLMConfInfo.TopLogProbs)
}
