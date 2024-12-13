# Secret Santa Bot 🎅🤖

@NotificationsSecret_bot
https://t.me/NotificationsSecret_bot

**Secret Santa Bot** — это Telegram-бот, который помогает организовать игру "Тайный Санта" в группах. Участники регистрируются через бота, после чего каждому случайным образом назначается "Тайный Санта".

## 📋 Возможности

- 🔹 **Регистрация участников**: Участники могут зарегистрироваться с помощью команды или кнопки.
- 🔹 **Назначение "Тайного Санты"**: Бот автоматически распределяет участников.
- 🔹 **Простое управление**: Полный список доступных команд доступен через `/help`.
- 🔹 **Поддержка групповых чатов**: Администраторы могут ограничивать использование команд.
- 🔹 **Сброс игры**: Возможность обнулить данные и начать заново.

## 🛠 Используемые технологии

- **Язык программирования**: [Go](https://golang.org)
- **Telegram API**: [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
- **База данных**: PostgreSQL (через [pgxpool](https://github.com/jackc/pgx))
- **Управление конфигурацией**: Переменные окружения и библиотека [godotenv](https://github.com/joho/godotenv)

## 🚀 Как запустить

### 🔧 Требования

- [Go](https://golang.org) 1.20 или выше
- PostgreSQL 12 или выше
- Файл `.env` с переменными окружения:

```dotenv
TELEGRAM_TOKEN=<токен вашего Telegram-бота>
DATABASE_URL=<URL подключения к PostgreSQL>
```

### 📥 Установка

1. Склонируйте репозиторий:
   ```bash
   git clone https://github.com/Ramcache/SecretSanta.git
   cd SecretSanta
   ```

2. Установите зависимости:
   ```bash
   go mod tidy
   ```

3. Настройте базу данных:
   Создайте таблицу `users`:
   ```sql
   CREATE TABLE users (
       id SERIAL PRIMARY KEY,
       telegram_id BIGINT NOT NULL,
       chat_id BIGINT NOT NULL,
       name TEXT NOT NULL,
       assigned BOOLEAN DEFAULT FALSE,
       UNIQUE (telegram_id, chat_id)
   );
   ```

4. Запустите бота:
   ```bash
   go run main.go
   ```

### 🐳 Запуск через Docker

1. Создайте Docker-образ:
   ```bash
   docker build -t secret-santa-bot .
   ```

2. Запустите контейнер:
   ```bash
   docker run --env-file .env secret-santa-bot
   ```

## 📚 Команды

- `/register_button` — отправить кнопку для регистрации участников.
- `/assign` — распределить "Тайных Сант" среди зарегистрированных участников.
- `/list_users` — показать список зарегистрированных участников.
- `/help` — показать список доступных команд.
- `/reset_game` — сбросить текущую игру и данные.

## 📂 Структура проекта

- **`main.go`**: Точка входа в приложение.
- **`config/`**: Настройка конфигурации.
- **`internal/db/`**: Работа с базой данных.
- **`internal/bot/`**: Основная логика бота.
- **`handlers.go`**: Обработка команд и сообщений.

## 🤝 Вклад в проект

Если хотите предложить улучшения или сообщить об ошибке, создайте issue или pull request. Мы рады любым предложениям!

## 📄 Лицензия

Проект распространяется под лицензией MIT.
