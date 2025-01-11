package db

import (
	"database/sql"
	"dt/utils/log"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

func NewAdapter(config *DatabaseConfig) (*sql.DB, error) {
	db, err := connectToDatabase(config)
	if err != nil {
		return nil, err
	}

	err = applyMigrations(config)
	if err != nil {
		return nil, err
	}

	log.Info("Database connection established successfully")
	return db, nil
}

func CloseConnection(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Error("Failed to close database connection")
	} else {
		log.Info("Database connection closed successfully")
	}
}

func connectToDatabase(config *DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable",
		config.Username, config.Password, config.DBName, config.Host, config.Port)
	var db *sql.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			break
		} else {
			log.Warn("Failed to connect to database %d, Retrying ...", i+1)
			time.Sleep(5 * time.Second)
		}
	}
	if err != nil {
		log.Error("Failed to connect to database")
		return nil, err
	}

	err = testDB(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func applyMigrations(config *DatabaseConfig) error {
	source, err := iofs.New(config.Migrations, config.MigrationsFolder)
	if err != nil {
		log.Error("Failed to create migrator source: %v", err)
		return fmt.Errorf("failed to create source: %w", err)
	}

	pgUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", config.Username, config.Password, config.Host, config.Port, config.DBName)
	migrator, err := migrate.NewWithSourceInstance("iofs", source, pgUrl)
	if err != nil {
		log.Error("Failed to create migrator: %v", err)
		return fmt.Errorf("migrate new: %s", err)
	}

	log.Debug("Migrations started")
	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error("Failed to apply migrations: %v", err)
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Debug("Migrations applied successfully")

	return nil
}

func testDB(db *sql.DB) error {
	err := db.Ping()
	if err != nil {
		return err
	}

	log.Info("Pinged database successfully!")
	return nil
}
