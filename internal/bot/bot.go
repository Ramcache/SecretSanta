package bot

import (
	"SecretSanta/internal/db"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

type Bot struct {
	TelegramBot *tgbotapi.BotAPI
	DB          *db.DB
	UserStates  map[int64]string // Состояния пользователей, ключ — Telegram ID
}

func NewBot(token string, database *db.DB) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		TelegramBot: bot,
		DB:          database,
		UserStates:  make(map[int64]string),
	}, nil

}

func (b *Bot) Run() {
	b.TelegramBot.Debug = true
	log.Printf("Authorized on account %s", b.TelegramBot.Self.UserName)

	updates := b.TelegramBot.GetUpdatesChan(tgbotapi.NewUpdate(0))

	for update := range updates {
		if update.Message != nil {
			// Проверяем состояние пользователя
			if state, exists := b.UserStates[update.Message.From.ID]; exists {
				log.Printf("User %d state: %s", update.Message.From.ID, state)
				if state == "awaiting_name" || strings.HasPrefix(state, "awaiting_name_") {
					b.handleNameInput(update.Message)
					delete(b.UserStates, update.Message.From.ID) // Сбрасываем состояние
				} else {
					b.sendReply(update.Message.Chat.ID, "Неподдерживаемое состояние.")
				}
			} else {
				// Обрабатываем команды
				b.HandleMessage(update.Message)
			}

		}
	}
}
