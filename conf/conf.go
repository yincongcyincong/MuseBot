package conf

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	BotToken      *string
	DeepseekToken *string
	DeepseekType  *string // simple complex
	CustomUrl     *string
	VolcAK        *string
	VolcSK        *string
	DBType        *string
	DBConf        *string
	
	AllowedTelegramUserIds = make(map[int64]bool)
)

func InitConf() {
	BotToken = flag.String("telegram_bot_token", "", "Comma-separated list of Telegram bot tokens")
	DeepseekToken = flag.String("deepseek_token", "", "deepseek auth token")
	CustomUrl = flag.String("custom_url", "https://api.deepseek.com/", "deepseek custom url")
	DeepseekType = flag.String("deepseek_type", "deepseek", "deepseek auth type")
	VolcAK = flag.String("volc_ak", "", "volc ak")
	VolcSK = flag.String("volc_sk", "", "volc sk")
	DBType = flag.String("db_type", "sqlite3", "db type")
	DBConf = flag.String("db_conf", "./data/telegram_bot.db", "db conf")
	allowedUserIds := flag.String("allowed_telegram_user_ids", "", "db conf")
	flag.Parse()

	if os.Getenv("TELEGRAM_BOT_TOKEN") != "" {
		*BotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	}

	if os.Getenv("DEEPSEEK_TOKEN") != "" {
		*DeepseekToken = os.Getenv("DEEPSEEK_TOKEN")
	}

	if os.Getenv("CUSTOM_URL") != "" {
		*CustomUrl = os.Getenv("CUSTOM_URL")
	}

	if os.Getenv("DEEPSEEK_TYPE") != "" {
		*DeepseekType = os.Getenv("DEEPSEEK_TYPE")
	}

	if os.Getenv("VOLC_AK") != "" {
		*VolcAK = os.Getenv("VOLC_AK")
	}

	if os.Getenv("VOLC_SK") != "" {
		*VolcSK = os.Getenv("VOLC_SK")
	}

	if os.Getenv("DB_TYPE") != "" {
		*DBType = os.Getenv("DB_TYPE")
	}

	if os.Getenv("DB_CONF") != "" {
		*DBConf = os.Getenv("DB_CONF")
	}

	if os.Getenv("ALLOWED_TELEGRAM_USER_IDS") != "" {
		*allowedUserIds = os.Getenv("ALLOWED_TELEGRAM_USER_IDS")
	}

	for _, userIdStr := range strings.Split(*allowedUserIds, ",") {
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			log.Printf("AllowedTelegramUserIds parse error:%s\n", userIdStr)
		}
		AllowedTelegramUserIds[int64(userId)] = true
	}

	fmt.Println("TelegramBotToken:", *BotToken)
	fmt.Println("DeepseekToken:", *DeepseekToken)
	fmt.Println("CustomUrl:", *CustomUrl)
	fmt.Println("DeepseekType:", *DeepseekType)
	fmt.Println("VOLC_AK:", *VolcAK)
	fmt.Println("VOLC_SK:", *VolcSK)
	fmt.Println("DBType:", *DBType)
	fmt.Println("DBConf:", *DBConf)
	fmt.Println("AllowedTelegramUserIds:", allowedUserIds)
	if *BotToken == "" || *DeepseekToken == "" {
		log.Fatalf("Bot token and deepseek token are required")
	}

}
