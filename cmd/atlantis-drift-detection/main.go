package main

import (
	"context"
	"fmt"
	"github.com/cresta/atlantis-drift-detection/internal/atlantis"
	"github.com/cresta/atlantis-drift-detection/internal/drifter"
	"github.com/cresta/atlantis-drift-detection/internal/notification"
	"github.com/cresta/atlantis-drift-detection/internal/resultcache"
	"github.com/cresta/atlantis-drift-detection/internal/terraform"
	"github.com/cresta/gogit"
	"github.com/cresta/gogithub"
	"github.com/joho/godotenv"
	"net/http"
	"os"

	// Empty import allows pinning to version atlantis uses
	_ "github.com/nlopes/slack"
	"go.uber.org/zap"
)
import "github.com/joeshaw/envdecode"

type config struct {
	Repo               string   `env:"REPO,required"`
	AtlantisHostname   string   `env:"ATLANTIS_HOST,required"`
	AtlantisToken      string   `env:"ATLANTIS_TOKEN,required"`
	DirectoryWhitelist []string `env:"DIRECTORY_WHITELIST"`
	SlackWebhookURL    string   `env:"SLACK_WEBHOOK_URL"`
	SkipWorkspaceCheck bool     `env:"SKIP_WORKSPACE_CHECK"`
	ParallelRuns       int      `env:"PARALLEL_RUNS"`
	DynamodbTable      string   `env:"DYNAMODB_TABLE"`
}

func loadEnvIfExists() error {
	_, err := os.Stat(".env")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error checking for .env file: %v", err)
	}
	return godotenv.Load()
}

type zapGogitLogger struct {
	logger *zap.Logger
}

func (z *zapGogitLogger) build(strings map[string]string, ints map[string]int64) *zap.Logger {
	l := z.logger
	for k, v := range strings {
		l = l.With(zap.String(k, v))
	}
	for k, v := range ints {
		l = l.With(zap.Int64(k, v))
	}
	return l
}

func (z *zapGogitLogger) Debug(_ context.Context, msg string, strings map[string]string, ints map[string]int64) {
	z.build(strings, ints).Debug(msg)
}

func (z *zapGogitLogger) Info(_ context.Context, msg string, strings map[string]string, ints map[string]int64) {
	z.build(strings, ints).Info(msg)
}

var _ gogit.Logger = (*zapGogitLogger)(nil)

func main() {
	ctx := context.Background()
	zapCfg := zap.NewProductionConfig()
	zapCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, err := zapCfg.Build(zap.AddCaller())
	if err != nil {
		panic(err)
	}
	if err := loadEnvIfExists(); err != nil {
		logger.Panic("Failed to load .env", zap.Error(err))
	}
	var cfg config
	if err := envdecode.Decode(&cfg); err != nil {
		logger.Panic("failed to decode config", zap.Error(err))
	}
	cloner := &gogit.Cloner{
		Logger: &zapGogitLogger{logger},
	}
	notif := &notification.Multi{
		Notifications: []notification.Notification{
			&notification.Zap{Logger: logger.With(zap.String("notification", "true"))},
		},
	}
	if cfg.SlackWebhookURL != "" {
		logger.Info("setting up slack webhook notification")
		notif.Notifications = append(notif.Notifications, &notification.SlackWebhook{
			WebhookURL: cfg.SlackWebhookURL,
			HTTPClient: http.DefaultClient,
		})
	}
	ghClient, err := gogithub.NewGQLClient(ctx, logger, nil)
	if err != nil {
		logger.Panic("failed to create github client", zap.Error(err))
	}
	tf := terraform.Client{
		Logger: logger.With(zap.String("terraform", "true")),
	}

	cache := resultcache.Noop{}
	if cfg.DynamodbTable != "" {
		logger.Info("setting up dynamodb result cache")
		cache = resultcache.NewDynamodb(cfg.DynamodbTable, logger)
	}

	d := drifter.Drifter{
		DirectoryWhitelist: cfg.DirectoryWhitelist,
		Logger:             logger.With(zap.String("drifter", "true")),
		Repo:               cfg.Repo,
		AtlantisClient: &atlantis.Client{
			AtlantisHostname: cfg.AtlantisHostname,
			Token:            cfg.AtlantisToken,
			HTTPClient:       http.DefaultClient,
		},
		ParallelRuns:       cfg.ParallelRuns,
		ResultCache:        cache,
		Cloner:             cloner,
		GithubClient:       ghClient,
		Terraform:          &tf,
		Notification:       notif,
		SkipWorkspaceCheck: cfg.SkipWorkspaceCheck,
	}
	if err := d.Drift(ctx); err != nil {
		logger.Panic("failed to drift", zap.Error(err))
	}
}
