package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"zeinai-backend/internal/config"
)

var (
	pool *pgxpool.Pool
	once sync.Once
)

// ConnectDB initializes the database connection pool
func ConnectDB(cfg config.Config) *pgxpool.Pool {
	once.Do(func() {
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.DB.Host,
			cfg.DB.Port,
			cfg.DB.User,
			cfg.DB.Password,
			cfg.DB.Name,
			cfg.DB.SSLMode,
		)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		config, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			log.Fatalf("Unable to parse DB config: %v\n", err)
		}

		// Pool configuration
		config.MaxConns = 25
		config.MinConns = 2
		config.MaxConnLifetime = 1 * time.Hour
		config.MaxConnIdleTime = 30 * time.Minute

		var errConnect error
		pool, errConnect = pgxpool.NewWithConfig(ctx, config)
		if errConnect != nil {
			log.Fatalf("Unable to connect to database: %v\n", errConnect)
		}

		if err := pool.Ping(ctx); err != nil {
			log.Fatalf("Unable to ping database: %v\n", err)
		}

		log.Println("Database connection established")
	})

	return pool
}

// GetDB returns the initialized db pool
func GetDB() *pgxpool.Pool {
	return pool
}

// CloseDB closes the database connection
func CloseDB() {
	if pool != nil {
		pool.Close()
	}
}
