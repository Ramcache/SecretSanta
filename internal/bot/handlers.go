package bot

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"math/rand/v2"
	"strconv"
	"strings"
)

var restrictedCommandsInGroups = map[string]bool{
	"start":  true,
	"assign": true,
}

func (b *Bot) HandleMessage(msg *tgbotapi.Message) {
	if !msg.IsCommand() {
		return // Игнорируем сообщения, которые не являются командами
	}

	// Ограничение для групп
	if msg.Chat.IsGroup() || msg.Chat.IsSuperGroup() {
		if restrictedCommandsInGroups[msg.Command()] {
			isAdmin, err := b.isUserAdmin(msg.From.ID, msg.Chat.ID)
			if err != nil {
				log.Printf("Error checking admin status: %v", err)
				b.sendReply(msg.Chat.ID, "Не удалось проверить права администратора. Попробуйте позже.")
				return
			}

			if !isAdmin {
				b.sendReply(msg.Chat.ID, "Эта команда доступна только администраторам группы.")
				return
			}
		}
	}

	// Обработка команд
	switch msg.Command() {
	case "start":
		b.handleStartCommand(msg)
	case "assign":
		b.handleAssignCommand(msg)
	case "register_button":
		b.handleRegisterButtonCommand(msg)
	case "list_users":
		b.handleListUsersCommand(msg)
	case "help":
		b.handleHelpCommand(msg)
	default:
		b.sendReply(msg.Chat.ID, "Неизвестная команда. Используйте /help для получения списка команд.")
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

func (b *Bot) isUserAdmin(userID int64, chatID int64) (bool, error) {
	// Создаем конфигурацию для получения администраторов
	config := tgbotapi.ChatAdministratorsConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: chatID,
		},
	}

	// Получаем список администраторов
	admins, err := b.TelegramBot.GetChatAdministrators(config)
	if err != nil {
		return false, err
	}

	// Проверяем, есть ли пользователь среди администраторов
	for _, admin := range admins {
		if admin.User.ID == userID {
			return true, nil
		}
	}

	return false, nil
}

func (b *Bot) handleHelpCommand(msg *tgbotapi.Message) {
	helpMessage := `
Добро пожаловать! Вот список доступных команд:

/register_button - Отправить кнопку регистрации в группу .
/assign - Назначить Тайного Санту.
/list_users - Показать список зарегистрированных пользователей.
/help - Показать это сообщение.

Если у вас возникли вопросы, обратитесь к администратору группы.
    `
	b.sendReply(msg.Chat.ID, helpMessage)
}
