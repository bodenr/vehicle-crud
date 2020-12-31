// Package db provides database specific implementation.
package db

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bodenr/vehicle-api/config"
	"github.com/bodenr/vehicle-api/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// singleton database instance
var (
	database *sqlx.DB
	lock     *sync.Mutex
)

func init() {
	lock = &sync.Mutex{}
}

// isConnectionError checks if the given error is connection refused.
func isConnectionError(err error) bool {
	// TODO: find a way to not check error string
	return strings.Contains(err.Error(), "connection refused")
}

// connect tries to connect to the database using the said dsn.
func connect(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Log.Err(err).Msg("error initializing database connection")
		return nil, err
	}
	return db, nil
}

// connectRetry retries connecting with the said dsn in cases of connection error.
func connectRetry(dsn string, retries int, delay time.Duration) (*sqlx.DB, error) {
	for retry := 0; retry < retries; retry++ {
		db, err := connect(dsn)
		if err == nil {
			return db, nil
		} else if isConnectionError(err) && delay > 0 {
			time.Sleep(delay)
		}
	}
	return nil, fmt.Errorf("Database connection failed after %d attempts", retries)
}

// Initialize should be called to initialize the database connection prior to GetDB.
func Initialize(conf *config.DatabaseConfig) error {
	lock.Lock()
	defer lock.Unlock()

	if database != nil {
		return fmt.Errorf("Database already initialized")
	}

	log.Log.Debug().Msg("initializing database connection")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		conf.Host, conf.Username, conf.Password, conf.DatabaseName, conf.Port, conf.Timezone)
	db, err := connect(dsn)

	// TODO: better handling to wait for DB up
	if err != nil && isConnectionError(err) && conf.ConnectRetries > 0 {
		db, err = connectRetry(dsn, conf.ConnectRetries, conf.ConnectBackoff)
		if err != nil {
			return err
		}
	}

	database = db
	log.Log.Debug().Msg("database initialized")

	return nil
}

// GetDB returns a reference to the singleton database, which maybe nil if not yet created via Initialize.
func GetDB() *sqlx.DB {
	if database == nil {
		log.Log.Warn().Msg("database accessed before initialization")
	}
	return database
}

// Close closes the database connection and clears the singleton database reference.
// This method is idempotent.
func Close() error {
	lock.Lock()
	defer lock.Unlock()

	if database == nil {
		log.Log.Warn().Msg("database already closed")
		return nil
	}
	if err := database.Close(); err != nil {
		log.Log.Err(err).Msg("error closing database")
		return err
	}
	database = nil
	log.Log.Debug().Msg("closed database")

	return nil
}
