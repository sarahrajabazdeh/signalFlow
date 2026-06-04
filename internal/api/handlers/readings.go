package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"signalflow/internal/models"
)

// HandleCreateReading submits a meter reading for async processing via NATS.
//
// @Summary      Submit a meter reading
// @Tags         readings
// @Accept       json
// @Param        reading  body  models.CreateReadingRequest  true  "Meter reading"
// @Success      202      "Accepted for async processing"
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /readings [post]
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
