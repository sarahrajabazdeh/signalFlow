package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"signalflow/internal/db/queries"
	"signalflow/internal/models"
)

func HandleCreateAsset(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.CreateAssetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
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
		json.NewEncoder(w).Encode(asset)
	}
}
