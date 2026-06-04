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
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		var req models.CreateReadingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.AssetID == uuid.Nil || req.ActualOutput < 0 || req.RecordedAt.IsZero() {
			writeError(w, http.StatusBadRequest, "asset_id, actual_output, and recorded_at are required")
			return
		}

		data, err := json.Marshal(req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		if _, err := js.Publish(subject, data); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to publish reading")
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
