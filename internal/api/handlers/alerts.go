package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"signalflow/internal/db/queries"
	"signalflow/internal/models"
)

func HandleListAlerts(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var assetID *uuid.UUID
		if raw := r.URL.Query().Get("asset_id"); raw != "" {
			id, err := uuid.Parse(raw)
			if err != nil {
				http.Error(w, "invalid asset_id", http.StatusBadRequest)
				return
			}
			assetID = &id
		}

		alerts, err := queries.ListAlerts(r.Context(), pool, assetID)
		if err != nil {
			http.Error(w, "failed to list alerts", http.StatusInternalServerError)
			return
		}
		if alerts == nil {
			alerts = []models.Alert{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(alerts)
	}
}
