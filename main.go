package main

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"go.albinodrought.com/creamy-prediction-market/internal/events"
	"go.albinodrought.com/creamy-prediction-market/internal/handlers"
	"go.albinodrought.com/creamy-prediction-market/internal/repo"
	"go.albinodrought.com/creamy-prediction-market/internal/types"
	"golang.org/x/crypto/bcrypt"
)

//go:embed ui/dist
var ui embed.FS

var logger = logrus.New()

type Config struct {
	Debug    bool   `json:"debug"`
	AdminPIN string `json:"admin_pin"`

	RepoPath string `json:"repo_path"`

	StartingTokens int64 `json:"starting_tokens"`
}

func main() {
	logger.SetFormatter(&logrus.JSONFormatter{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gracefulCtx, gracefulCancel := context.WithCancel(ctx)

	var (
		configBytes []byte
		err         error
	)

	env := os.Getenv("CREAMY_PM_CONFIG")

	if env == "" {
		configBytes, err = os.ReadFile("config.json")
		if err != nil {
			logger.WithError(err).Fatal("failed to read config.json")
		}
	} else {
		configBytes = []byte(env)
	}

	config := Config{}
	if err = json.Unmarshal(configBytes, &config); err != nil {
		logger.WithError(err).WithField("raw-config", string(configBytes)).Fatal("failed to unmarshal config")
	}

	if config.Debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	store := repo.NewStore()

	(func() {
		if config.RepoPath != "" {
			handle, err := os.Open(config.RepoPath)
			if err == nil {
				defer handle.Close()
				err = store.Load(handle)
				if err != nil {
					logger.WithError(err).Fatal("failed to load repo from path")
				}
				logger.Info("loaded repo")
				return
			}

			if !os.IsNotExist(err) {
				logger.WithError(err).Fatal("unhandled error while opening repo path")
			}

			// file doesn't exist, proceed with default data creation
		}

		adminID, err := repo.NewID()
		if err != nil {
			logger.WithError(err).Fatal("failed to generate admin ID")
		}

		adminPin := config.AdminPIN
		if adminPin == "" {
			adminPin = "0000"
		}

		pinHash, err := bcrypt.GenerateFromPassword([]byte(adminPin), bcrypt.MinCost)
		if err != nil {
			logger.WithError(err).Fatal("failed to generate admin bcrypt hash")
		}

		err = store.AddUser(types.User{
			ID:      adminID,
			Name:    "Admin",
			PINHash: pinHash,
			Admin:   true,
			Tokens:  0,
		}, 0)
		if err != nil {
			logger.WithError(err).Fatal("failed to add admin user")
		}
		logger.Info("added admin user")
	})()

	// Create and start event hub for SSE
	eventHub := events.NewHub()
	go eventHub.Run()

	// save repo data every minute if dirty
	save := func() {
		if !store.IsDirty() {
			return
		}

		newPath := config.RepoPath + ".new"
		oldPath := config.RepoPath + ".old"

		handle, err := os.Create(newPath)
		if err != nil {
			logger.WithError(err).Warn("failed to create persistence path!")
			return
		}
		defer handle.Close()

		err1 := store.Save(handle)
		if err1 != nil {
			logger.WithError(err1).Warn("failed to save store to handle")
		}
		err2 := handle.Close()
		if err2 != nil {
			logger.WithError(err2).Warn("failed to close handle")
		}
		if err1 != nil || err2 != nil {
			return
		}

		err = os.Rename(config.RepoPath, oldPath)
		if err != nil && !os.IsNotExist(err) {
			logger.WithError(err).Warn("failed to backup repo data to .old, ignoring")
		}
		err = os.Rename(newPath, config.RepoPath)
		if err != nil {
			logger.WithError(err).Warn("failed to move .new repo data to path!")
		} else {
			logger.Info("saved state")
		}
	}
	go func() {
		if config.RepoPath == "" {
			logger.Warn("running without persistence, please configure repo_path!")
			return
		}

		ticker := time.NewTicker(time.Minute)
		for {
			<-ticker.C
			save()
		}
	}()

	go func() {
		c := make(chan os.Signal, 2)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
		<-c
		logger.Info("received signal, performing graceful exit")
		gracefulCancel()
	}()

	h := &handlers.Handler{
		GracefulCtx:    gracefulCtx,
		Store:          store,
		Logger:         logger,
		StartingTokens: config.StartingTokens,
		EventHub:       eventHub,
	}

	// Sweep expired predictions every minute
	go func() {
		// Run once at startup
		h.Sweep()
		ticker := time.NewTicker(time.Minute)
		for {
			<-ticker.C
			h.Sweep()
		}
	}()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	sub, err := fs.Sub(ui, "ui/dist")
	if err != nil {
		logger.WithError(err).Fatal("failed to move into ui/dist subfolder of embedded UI bundle")
	}

	for _, spaRoute := range []string{
		"/home",
		"/predictions/{id}",
		"/leaderboard",
		"/my-bets",
		"/admin",
		"/admin/predictions/new",
		"/admin/predictions/{id}",
		"/admin/users",
	} {
		mux.Handle("GET "+spaRoute, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFileFS(w, r, sub, "index.html")
		}))
	}

	mux.Handle("GET /", http.FileServerFS(sub))

	server := http.Server{
		Addr:    ":3000",
		Handler: mux,
	}
	serverCtx, serverCancel := context.WithCancel(ctx)
	go func() {
		defer serverCancel()
		logger.Info("starting server on :3000")
		if err := server.ListenAndServe(); err != nil {
			logger.WithError(err).Error("http.ListenAndServe error")
		}
	}()

	select {
	case <-serverCtx.Done():
		logger.Info("received forceful shutdown")
	case <-gracefulCtx.Done():
		logger.Info("received graceful shutdown")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	server.Shutdown(shutdownCtx)

	save()

	logger.Info("goodnight")
}
