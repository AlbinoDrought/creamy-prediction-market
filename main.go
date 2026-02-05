package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

type Config struct {
	Debug bool `json:"debug"`

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

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})

	if err := http.ListenAndServe(":3000", mux); err != nil {
		logger.WithError(err).Error("http.ListenAndServe error")
	}
	logger.Info("end of main")
}
