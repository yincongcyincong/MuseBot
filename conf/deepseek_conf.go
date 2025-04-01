package conf

import (
	"flag"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"os"
	"strconv"
	"strings"
)

var (
	FrequencyPenalty *float64
	MaxTokens        *int
	PresencePenalty  *float64
	Temperature      *float64
	TopP             *float64
	Stop             []string
	LogProbs         *bool
	TopLogProbs      *int
)

func InitDeepseekConf() {
	FrequencyPenalty = flag.Float64("frequency_penalty", 0.0, "frequency penalty")
	MaxTokens = flag.Int("max_tokens", 2048, "maximum number of tokens")
	PresencePenalty = flag.Float64("presence_penalty", 0.0, "presence penalty")
	Temperature = flag.Float64("temperature", 0.7, "temperature")
	TopP = flag.Float64("top_p", 0.9, "top p")
	LogProbs = flag.Bool("log_probs", false, "log probs")
	TopLogProbs = flag.Int("top_log_probs", 0, "number of top log probs to return")

	stop := flag.String("stop", "", "stop sequence")

	if os.Getenv("FREQUENCY_PENALTY") != "" {
		*FrequencyPenalty, _ = strconv.ParseFloat(os.Getenv("FREQUENCY_PENALTY"), 64)
	}

	if os.Getenv("MAX_TOKENS") != "" {
		*MaxTokens, _ = strconv.Atoi(os.Getenv("MAX_TOKENS"))
	}

	if os.Getenv("PRESENCE_PENALTY") != "" {
		*PresencePenalty, _ = strconv.ParseFloat(os.Getenv("PRESENCE_PENALTY"), 64)
	}

	if os.Getenv("TEMPERATURE") != "" {
		*Temperature, _ = strconv.ParseFloat(os.Getenv("TEMPERATURE"), 64)
	}

	if os.Getenv("TOP_P") != "" {
		*TopP, _ = strconv.ParseFloat(os.Getenv("TOP_P"), 64)
	}

	if os.Getenv("STOP") != "" {
		*stop = os.Getenv("STOP")
	}

	if os.Getenv("LOG_PROBS") != "" {
		*LogProbs, _ = strconv.ParseBool(os.Getenv("LOG_PROBS"))
	}

	if os.Getenv("TOP_LOG_PROBS") != "" {
		*TopLogProbs, _ = strconv.Atoi(os.Getenv("TOP_LOG_PROBS"))
	}

	for _, s := range strings.Split(*stop, ",") {
		if s != "" {
			Stop = append(Stop, s)
		}

	}

	logger.Info("DEEPSEEK_CONF", "FrequencyPenalty", *FrequencyPenalty)
	logger.Info("DEEPSEEK_CONF", "MaxTokens", *MaxTokens)
	logger.Info("DEEPSEEK_CONF", "PresencePenalty", *PresencePenalty)
	logger.Info("DEEPSEEK_CONF", "Temperature", *Temperature)
	logger.Info("DEEPSEEK_CONF", "TopP", *TopP)
	logger.Info("DEEPSEEK_CONF", "Stop", *stop)
	logger.Info("DEEPSEEK_CONF", "LogProbs", *LogProbs)
	logger.Info("DEEPSEEK_CONF", "TopLogProbs", *TopLogProbs)
}
