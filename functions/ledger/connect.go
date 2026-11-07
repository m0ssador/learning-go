package ledger

import (
	"errors"
	"log"

	"github.com/m0ssador/learning-go/functions/ledger/internal/cache"
	"github.com/m0ssador/learning-go/functions/ledger/internal/db"
)

var errNoDatabase = errors.New("database is not connected")

func Connect() error {
	if err := db.Open(); err != nil {
		return err
	}
	log.Printf("подключение к PostgreSQL установлено")
	if err := cache.Connect(); err != nil {
		_ = db.Close()
		return err
	}
	return nil
}

func Connected() bool {
	return db.SQL() != nil
}
