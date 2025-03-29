package component

import (
	"context"
	"fmt"
	"time"

	shared "notify-service/internal"

	log "github.com/sirupsen/logrus"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"go.uber.org/fx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDb(
	lc fx.Lifecycle,
	config *shared.Config,
) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable search_path=%s",
		config.PostgresHost,
		config.PostgresUser,
		config.PostgresPassword,
		config.PostgresDb,
		config.PostgresPort,
		config.PostgresSchema,
	)
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement cache
	}), &gorm.Config{})
	if err != nil {
		log.WithError(err).Fatal("error connecting to database")
	}
	err = db.Use(otelgorm.NewPlugin())
	if err != nil {
		log.WithError(err).Fatal("error using otelgorm plugin")
	}
	// Config the connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.WithError(err).Fatal("error getting db connection")
	}
	sqlDB.SetMaxIdleConns(config.DbMaxIdleConns)
	sqlDB.SetMaxOpenConns(config.DbMaxOpenConns)
	sqlDB.SetConnMaxIdleTime(time.Minute)
	sqlDB.SetConnMaxLifetime(time.Hour)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Shutting down database connection...")
			return sqlDB.Close()
		},
	})
	return db
}
