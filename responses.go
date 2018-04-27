package sabakan

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/log"
)

func renderJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		renderError(w, err, http.StatusInternalServerError)
	}
}

func renderError(ctx context.Context, w http.ResponseWriter, e APIError) {
	fields := cmd.FieldsFromContext(ctx)
	fields["status"] = e.Status
	fields[log.FnError] = e.Error()
	log.Error(http.StatusText(e.Status), fields)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	err := json.NewEncoder(w).Encode(out)
	if err != nil {
		log.Error("failed to output JSON", map[string]interface{}{
			log.FnError: err.Error(),
		})
	}
}
