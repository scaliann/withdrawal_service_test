package render

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

type Err struct {
	Error string `json:"error"`
}

func Error(w http.ResponseWriter, err error, status int, message string) {
	if err != nil {
		log.Error().
			Err(err).
			Int("status", status).
			Str("message", message).
			Msg("request failed")
	}

	JSON(w, Err{Error: message}, status)
}
