package conf

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	BotToken      *string
	DeepseekToken *string
)

func InitConf() {
	BotToken = flag.String("telegram_bot_token", "", "Comma-separated list of Telegram bot tokens")
	DeepseekToken = flag.String("deepseek_token", "", "deepseek auth token")
	flag.Parse()

	if *BotToken == "" {
		*BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	}

	if *DeepseekToken == "" {
		*DeepseekToken = os.Getenv("DEEPSEEK_TOKEN")
	}

	fmt.Println("TelegramBotToken:", *BotToken)
	fmt.Println("DeepseekToken:", *DeepseekToken)
	if *BotToken == "" || *DeepseekToken == "" {
		log.Fatalf("Bot token and deepseek token are required")
	}

}
