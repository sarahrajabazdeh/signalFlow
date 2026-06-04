package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"signalflow/internal/db/queries"
	"signalflow/internal/models"
)

func HandleCreateAsset(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
		var req models.CreateAssetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" || req.ExpectedOutput <= 0 {
			http.Error(w, "name and expected_output are required", http.StatusBadRequest)
			return
		}

		asset := models.Asset{
			ID:             uuid.New(),
			Name:           req.Name,
			ExpectedOutput: req.ExpectedOutput,
			CreatedAt:      time.Now().UTC(),
		}

		if err := queries.SaveAsset(r.Context(), pool, asset); err != nil {
			http.Error(w, "failed to save asset", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(asset); err != nil {
			log.Error().Err(err).Msg("encode response")
		}
	}
}
