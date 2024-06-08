package config

import (
	"database/sql"
	"log"
)

type Config struct {
	Channels []string
	Keywords []string
}

func LoadConfigFromDB(db *sql.DB) (*Config, error) {
	config := &Config{}

	// Загрузка каналов
	rows, err := db.Query("SELECT name FROM channels")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			log.Fatal(err)
		}
		config.Channels = append(config.Channels, name)
	}

	// Загрузка ключевых слов
	rows, err = db.Query("SELECT keyword FROM keywords")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var keyword string
		if err := rows.Scan(&keyword); err != nil {
			log.Fatal(err)
		}
		config.Keywords = append(config.Keywords, keyword)
	}

	return config, nil
}
