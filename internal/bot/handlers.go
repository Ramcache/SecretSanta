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
	case "list_users":
		b.handleListUsersCommand(msg)
	default:
		b.sendReply(msg.Chat.ID, "Неизвестная команда. Используйте /start, /register, /assign, /register_button.")
	}
}

func (b *Bot) handleStartCommand(msg *tgbotapi.Message) {
	args := msg.CommandArguments()

	if strings.HasPrefix(args, "register_group_") {
		// Извлекаем ID группы
		groupID, err := strconv.ParseInt(strings.TrimPrefix(args, "register_group_"), 10, 64)
		if err != nil {
			b.sendReply(msg.Chat.ID, "Ошибка: некорректный ID группы.")
			log.Printf("Failed to parse group ID: %v", err)
			return
		}

		// Сохраняем ID группы в состоянии пользователя
		b.UserStates[msg.From.ID] = fmt.Sprintf("awaiting_name_%d", groupID)
		log.Printf("Set state for user %d: awaiting_name_%d", msg.From.ID, groupID)
		b.sendReply(msg.Chat.ID, "Пожалуйста, напишите своё имя для регистрации.")
		return
	}

	b.sendReply(msg.Chat.ID, "Добро пожаловать! Используйте команды бота.")
}

func (b *Bot) handleRegisterCommand(msg *tgbotapi.Message) {
	// Устанавливаем состояние для пользователя
	b.UserStates[msg.From.ID] = "awaiting_name"

	// Отправляем сообщение с просьбой ввести имя
	b.sendReply(msg.Chat.ID, "Пожалуйста, напишите ваше имя для регистрации.")
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

	// Очистка базы данных для текущего чата
	err = b.DB.ClearUsersForChat(msg.Chat.ID)
	if err != nil {
		log.Printf("Ошибка при очистке базы данных для группы %d: %v", msg.Chat.ID, err)
		b.sendReply(msg.Chat.ID, "Ошибка при очистке базы данных. Пожалуйста, обратитесь к администратору.")
	} else {
		log.Printf("Участники для группы %d успешно удалены из базы данных.", msg.Chat.ID)
	}
}

func (b *Bot) sendReply(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := b.TelegramBot.Send(msg)
	if err != nil {
		log.Printf("Failed to send message to chatID %d: %v", chatID, err)
	}
}

func (b *Bot) handleRegisterButtonCommand(msg *tgbotapi.Message) {
	botUsername := b.TelegramBot.Self.UserName
	groupID := msg.Chat.ID // ID группы, откуда нажимается кнопка

	// Создаем кнопку с передачей ID группы
	registerButton := tgbotapi.NewInlineKeyboardButtonURL(
		"Зарегистрироваться",
		fmt.Sprintf("https://t.me/%s?start=register_group_%d", botUsername, groupID),
	)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(registerButton))

	// Отправляем сообщение с кнопкой
	reply := tgbotapi.NewMessage(msg.Chat.ID, "Для регистрации нажмите кнопку ниже.")
	reply.ReplyMarkup = keyboard

	if _, err := b.TelegramBot.Send(reply); err != nil {
		log.Printf("Failed to send registration button: %v", err)
	}
}

func (b *Bot) handleNameInput(msg *tgbotapi.Message) {
	// Проверяем состояние пользователя
	state, exists := b.UserStates[msg.From.ID]
	if !exists || !strings.HasPrefix(state, "awaiting_name_") {
		b.sendReply(msg.Chat.ID, "Ошибка: ваше имя не ожидается. Используйте /register.")
		log.Printf("State not found or unsupported for user %d: %v", msg.From.ID, state)
		return
	}

	// Извлекаем ID группы из состояния
	groupID, err := strconv.ParseInt(strings.TrimPrefix(state, "awaiting_name_"), 10, 64)
	if err != nil {
		b.sendReply(msg.Chat.ID, "Ошибка: некорректный ID группы.")
		log.Printf("Failed to parse group ID from state: %v", err)
		return
	}

	// Регистрируем пользователя
	err = b.DB.AddUser(msg.From.ID, groupID, msg.Text)
	if err != nil {
		b.sendReply(msg.Chat.ID, "Ошибка при регистрации. Попробуйте позже.")
		log.Printf("Error registering user: %v", err)
		return
	}

	// Отправляем подтверждение
	b.sendReply(msg.Chat.ID, fmt.Sprintf("Вы успешно зарегистрировались как %s!", msg.Text))

	// Отправляем сообщение в группу
	announcement := fmt.Sprintf("Пользователь %s успешно зарегистрировался!", msg.Text)
	b.sendReply(groupID, announcement)

	// Сбрасываем состояние
	delete(b.UserStates, msg.From.ID)
}

func (b *Bot) handleListUsersCommand(msg *tgbotapi.Message) {
	users, err := b.DB.GetUnassignedUsers(msg.Chat.ID)
	if err != nil {
		b.sendReply(msg.Chat.ID, "Ошибка при получении списка пользователей.")
		log.Printf("Error fetching users: %v", err)
		return
	}

	if len(users) == 0 {
		b.sendReply(msg.Chat.ID, "Нет доступных пользователей.")
		return
	}

	reply := "Список участников:\n"
	for _, user := range users {
		reply += fmt.Sprintf("- %s (Telegram ID: %d)\n", user.Name, user.TelegramID)
	}

	b.sendReply(msg.Chat.ID, reply)
}
