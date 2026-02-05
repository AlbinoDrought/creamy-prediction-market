package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
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

	StartingTokens int64 `json:"starting_tokens"`
}

func main() {
	logger.SetFormatter(&logrus.JSONFormatter{})

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

	// Create admin user
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

	// Create sample prediction for testing
	if config.Debug {
		predictionID, err := repo.NewID()
		if err != nil {
			logger.WithError(err).Warn("failed to generate prediction ID, skipping")
		} else {
			choiceID1, _ := repo.NewID()
			choiceID2, _ := repo.NewID()
			choiceID3, _ := repo.NewID()

			err = store.PutPrediction(types.Prediction{
				ID:          predictionID,
				CreatedAt:   time.Now().Format(time.RFC3339),
				Name:        "What color will the balloon be?",
				Description: "The balloon will be released at noon",
				Status:      types.PredictionStatusOpen,
				ClosesAt:    time.Now().AddDate(0, 0, 1).Truncate(time.Hour).Add(11 * time.Hour).Format(time.RFC3339),
				Choices: []types.PredictionChoice{
					{ID: choiceID1, Name: "Red"},
					{ID: choiceID2, Name: "Green"},
					{ID: choiceID3, Name: "Blue"},
				},
				OddsVisibleBeforeBet: true,
			})
			if err != nil {
				logger.WithError(err).Warn("failed to add sample prediction")
			} else {
				logger.Info("added sample prediction")
			}
		}
	}

	// Create and start event hub for SSE
	eventHub := events.NewHub()
	go eventHub.Run()

	h := &handlers.Handler{
		Store:          store,
		Logger:         logger,
		StartingTokens: config.StartingTokens,
		EventHub:       eventHub,
	}

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

	logger.Info("starting server on :3000")
	if err := http.ListenAndServe(":3000", mux); err != nil {
		logger.WithError(err).Error("http.ListenAndServe error")
	}
	logger.Info("end of main")
}
