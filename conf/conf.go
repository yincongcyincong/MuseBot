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
	Mode          *string // simple complex
	CustomUrl     *string
)

func InitConf() {
	BotToken = flag.String("telegram_bot_token", "", "Comma-separated list of Telegram bot tokens")
	DeepseekToken = flag.String("deepseek_token", "", "deepseek auth token")
	CustomUrl = flag.String("custom_url", "https://api.deepseek.com/", "deepseek custom url")
	flag.Parse()

	if *BotToken == "" {
		*BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	}

	if *DeepseekToken == "" {
		*DeepseekToken = os.Getenv("DEEPSEEK_TOKEN")
	}

	if *CustomUrl == "" {
		*CustomUrl = os.Getenv("CUSTOM_URL")
	}

	fmt.Println("TelegramBotToken:", *BotToken)
	fmt.Println("DeepseekToken:", *DeepseekToken)
	fmt.Println("CustomUrl:", *CustomUrl)
	if *BotToken == "" || *DeepseekToken == "" {
		log.Fatalf("Bot token and deepseek token are required")
	}

}
