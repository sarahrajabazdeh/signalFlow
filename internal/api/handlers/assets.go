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

// HandleCreateAsset creates a new energy asset.
//
// @Summary      Create an asset
// @Tags         assets
// @Accept       json
// @Produce      json
// @Param        asset  body      models.CreateAssetRequest  true  "Asset to create"
// @Success      201    {object}  models.Asset
// @Failure      400    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /assets [post]
func HandleCreateAsset(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req models.CreateAssetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Name == "" || req.ExpectedOutput <= 0 {
			writeError(w, http.StatusBadRequest, "name and expected_output are required")
			return
		}

		asset := models.Asset{
			ID:             uuid.New(),
			Name:           req.Name,
			ExpectedOutput: req.ExpectedOutput,
			CreatedAt:      time.Now().UTC(),
		}
		if err := queries.SaveAsset(r.Context(), pool, asset); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save asset")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(asset); err != nil {
			log.Error().Err(err).Msg("encode response")
		}
	}
}
