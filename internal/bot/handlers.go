package bot

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"math/rand/v2"
	"strconv"
	"strings"
)

func (b *Bot) HandleMessage(msg *tgbotapi.Message) {
	if !msg.IsCommand() {
		return
	}

	switch msg.Command() {
	case "start":
		b.handleStartCommand(msg)
	case "register":
		b.handleRegisterCommand(msg)
	case "assign":
		b.handleAssignCommand(msg)
	case "register_button":
		b.handleRegisterButtonCommand(msg)
	default:
		b.sendReply(msg.Chat.ID, "Неизвестная команда. Используйте /start, /register, /assign, /register_button.")
	}
}

func (b *Bot) handleStartCommand(msg *tgbotapi.Message) {
	args := msg.CommandArguments()

	var groupID int64
	if args != "" && strings.HasPrefix(args, "group_") {
		parsedID, err := strconv.ParseInt(strings.TrimPrefix(args, "group_"), 10, 64)
		if err == nil {
			groupID = parsedID
		} else {
			log.Printf("Failed to parse group ID: %v", err)
		}
	}

	if groupID != 0 {
		reply := fmt.Sprintf("Вы пришли из группы с ID %d. Вернитесь, чтобы продолжить игру.", groupID)

		groupLink := fmt.Sprintf("https://t.me/c/%d", groupID) // Пример ссылки для возврата
		returnButton := tgbotapi.NewInlineKeyboardButtonURL("Вернуться в группу", groupLink)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(returnButton))

		msg := tgbotapi.NewMessage(msg.Chat.ID, reply)
		msg.ReplyMarkup = keyboard

		if _, err := b.TelegramBot.Send(msg); err != nil {
			log.Printf("Failed to send return button: %v", err)
		}
	} else {
		// Если параметр не передан, показываем стандартное сообщение
		b.sendReply(msg.Chat.ID, "Добро пожаловать в Тайный Санта! Перейдите в группу для регистрации.")
	}
}

func (b *Bot) handleRegisterCommand(msg *tgbotapi.Message) {
	args := msg.CommandArguments()
	if args == "" {
		b.sendReply(msg.Chat.ID, "Пожалуйста, укажите имя после команды /register.")
		return
	}

	err := b.DB.AddUser(msg.From.ID, msg.Chat.ID, args)
	if err != nil {
		b.sendReply(msg.Chat.ID, "Не удалось зарегистрироваться. Попробуйте позже.")
		log.Printf("Error registering user: %v", err)
		return
	}

	b.sendReply(msg.Chat.ID, fmt.Sprintf("Вы успешно зарегистрировались как %s!", args))
}

func (b *Bot) handleAssignCommand(msg *tgbotapi.Message) {
	users, err := b.DB.GetUnassignedUsers(msg.Chat.ID)
	if err != nil {
		b.sendReply(msg.Chat.ID, "Ошибка при назначении Тайного Санты.")
		log.Printf("Error querying users: %v", err)
		return
	}

	if len(users) < 2 {
		b.sendReply(msg.Chat.ID, "Недостаточно участников для игры.")
		return
	}

	// Перемешиваем список участников
	rand.Shuffle(len(users), func(i, j int) {
		users[i], users[j] = users[j], users[i]
	})

	// Создаём распределение
	assignments := make(map[string]string)
	for i := range users {
		assignTo := (i + 1) % len(users) // Следующий участник (циклически)
		assignments[users[i].Name] = users[assignTo].Name
		err := b.DB.MarkUserAssigned(users[i].ID)
		if err != nil {
			log.Printf("Error updating user assignment: %v", err)
			continue
		}
	}

	// Отправляем каждому участнику его назначение
	for _, user := range users {
		santaName, ok := assignments[user.Name]
		if !ok {
			log.Printf("Ошибка: для пользователя %s не найден Тайный Санта", user.Name)
			continue
		}
		reply := fmt.Sprintf("Ваш Тайный Санта: %s!", santaName)
		b.sendReply(user.TelegramID, reply)
	}

	b.sendReply(msg.Chat.ID, "Назначения завершены! Каждый получил имя своего Тайного Санты через бота.")
}

func (b *Bot) sendReply(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := b.TelegramBot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message to chatID %d: %v", chatID, err)
	}
}

func (b *Bot) handleRegisterButtonCommand(msg *tgbotapi.Message) {
	botUsername := b.TelegramBot.Self.UserName // Получаем имя бота

	// Создаем кнопку с ссылкой на бота
	registerButton := tgbotapi.NewInlineKeyboardButtonURL("Регистрация", fmt.Sprintf("https://t.me/%s?start=register", botUsername))
	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(registerButton))

	// Отправляем сообщение с кнопкой
	reply := tgbotapi.NewMessage(msg.Chat.ID, "Для регистрации, нажмите на кнопку ниже и запустите бота. После запуска вернитесь в групп и напишите /register <ваше имя>")
	reply.ReplyMarkup = keyboard

	if _, err := b.TelegramBot.Send(reply); err != nil {
		log.Printf("Failed to send registration button: %v", err)
	}
}
