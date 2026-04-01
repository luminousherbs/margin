package api

import (
	"context"
	"encoding/json"
	"net/http"

	"margin.at/internal/xrpc"
)

func (h *Handler) SyncAll(w http.ResponseWriter, r *http.Request) {
	session, err := h.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		WriteUnauthorized(w, err.Error())
		return
	}

	results, err := h.syncService.PerformSync(r.Context(), session.DID, func(ctx context.Context, did string) (*xrpc.Client, error) {
		var client *xrpc.Client
		err := h.refresher.ExecuteWithAutoRefresh(r, session, func(c *xrpc.Client, d string) error {
			client = c
			return nil
		})
		return client, err
	})

	if err != nil {
		WriteInternalError(w, "Sync failed")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}
