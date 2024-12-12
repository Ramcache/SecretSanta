package db

import (
	"context"
)

type User struct {
	ID         int
	TelegramID int64
	Name       string
}

func (db *DB) AddUser(telegramID int64, chatID int64, name string) error {
	_, err := db.Pool.Exec(context.Background(),
		"INSERT INTO users (telegram_id, chat_id, name) VALUES ($1, $2, $3) ON CONFLICT (telegram_id, chat_id) DO UPDATE SET name = $3",
		telegramID, chatID, name)
	return err
}

func (db *DB) GetUnassignedUsers(chatID int64) ([]User, error) {
	rows, err := db.Pool.Query(context.Background(), "SELECT id, telegram_id, name FROM users WHERE chat_id = $1 AND assigned = false", chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.TelegramID, &user.Name); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (db *DB) MarkUserAssigned(userID int) error {
	_, err := db.Pool.Exec(context.Background(), "UPDATE users SET assigned = true WHERE id = $1", userID)
	return err
}
