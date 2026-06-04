package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"signalflow/internal/models"
)

func HandleCreateReading(js nats.JetStreamContext, subject string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
		var req models.CreateReadingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.AssetID == uuid.Nil || req.ActualOutput < 0 || req.RecordedAt.IsZero() {
			http.Error(w, "asset_id, actual_output, and recorded_at are required", http.StatusBadRequest)
			return
		}

		data, err := json.Marshal(req)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		if _, err := js.Publish(subject, data); err != nil {
			http.Error(w, "failed to publish reading", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
