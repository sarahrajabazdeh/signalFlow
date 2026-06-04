package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"signalflow/internal/db/queries"
	"signalflow/internal/models"
)

// HandleListAlerts returns all alerts, optionally filtered by asset.
//
// @Summary      List imbalance alerts
// @Tags         alerts
// @Produce      json
// @Param        asset_id  query     string  false  "Filter by asset UUID"
// @Success      200       {array}   models.Alert
// @Failure      400       {object}  map[string]string
// @Failure      500       {object}  map[string]string
// @Router       /alerts [get]
func HandleListAlerts(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var assetID *uuid.UUID
		if raw := r.URL.Query().Get("asset_id"); raw != "" {
			id, err := uuid.Parse(raw)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid asset_id")
				return
			}
			assetID = &id
		}

		alerts, err := queries.ListAlerts(r.Context(), pool, assetID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to list alerts")
			return
		}
		if alerts == nil {
			alerts = []models.Alert{}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(alerts); err != nil {
			log.Error().Err(err).Msg("encode response")
		}
	}
}
