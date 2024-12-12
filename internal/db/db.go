package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func InitDB(databaseURL string) (*DB, error) {
	dbpool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, err
	}

	return &DB{Pool: dbpool}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}

func (db *DB) ClearUsersForChat(chatID int64) error {
	_, err := db.Pool.Exec(context.Background(), "DELETE FROM users WHERE chat_id = $1", chatID)
	return err
}
