package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/zelenin/go-tdlib/client"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"telejob/internal/config"
	"telejob/internal/db"
	"telejob/internal/process"
)

func main() {
	database := initializeDatabase()
	defer database.Close()

	tdClient := initializeTelegramClient()

	_, err := tdClient.GetMe()
	if err != nil {
		log.Fatalf("Failed to get me: %v", err)
	}

	cfg := loadConfiguration(database)
	process.ProcessChannels(cfg, tdClient, database)
}

func initializeDatabase() *sql.DB {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	database, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	db.ApplyMigrations(database, "postgres", "migrations")
	return database
}

func initializeTelegramClient() *client.Client {
	authorizer := client.ClientAuthorizer()
	go client.CliInteractor(authorizer)

	apiIdString := os.Getenv("TELEGRAM_API_ID")
	apiHash := os.Getenv("TELEGRAM_API_HASH")

	if apiIdString == "" || apiHash == "" {
		log.Fatal("TELEGRAM_API_ID and TELEGRAM_API_HASH environment variables must be set")
	}

	apiId, err := strconv.Atoi(apiIdString)
	if err != nil {
		log.Fatalf("Invalid TELEGRAM_API_ID: %v", err)
	}

	authorizer.TdlibParameters <- &client.SetTdlibParametersRequest{
		UseTestDc:           false,
		DatabaseDirectory:   filepath.Join(".tdlib", "database"),
		FilesDirectory:      filepath.Join(".tdlib", "files"),
		UseFileDatabase:     false,
		UseChatInfoDatabase: false,
		UseMessageDatabase:  false,
		UseSecretChats:      false,
		ApiId:               int32(apiId),
		ApiHash:             apiHash,
		SystemLanguageCode:  "en",
		DeviceModel:         "Server",
		SystemVersion:       "1.0.0",
		ApplicationVersion:  "1.0.0",
	}

	_, err = client.SetLogVerbosityLevel(&client.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: 1,
	})
	if err != nil {
		log.Fatalf("SetLogVerbosityLevel error: %v", err)
	}

	tdlibClient, err := client.NewClient(authorizer)
	if err != nil {
		log.Fatalf("NewClient error: %v", err)
	}

	optionValue, err := client.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		log.Fatalf("GetOption error: %v", err)
	}

	log.Printf("TDLib version: %s", optionValue.(*client.OptionValueString).Value)

	me, err := tdlibClient.GetMe()
	if err != nil {
		log.Fatalf("GetMe error: %v", err)
	}

	log.Printf("Me: %s %s [%s]", me.FirstName, me.LastName, me.Usernames)

	return tdlibClient
}

func loadConfiguration(database *sql.DB) *config.Config {
	cfg, err := config.LoadConfigFromDB(database)
	if err != nil {
		log.Fatalf("Failed to load config from DB: %v", err)
	}
	return cfg
}
