package drifter

import (
	"context"
	"encoding/json"
	"github.com/cresta/atlantis-drift-detection/internal/atlantis"
	"github.com/cresta/atlantis-drift-detection/internal/notification"
	"github.com/cresta/atlantis-drift-detection/internal/processedcache"
	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func lockedPlanResponse(t *testing.T) []byte {
	body, err := json.Marshal(command.Result{
		ProjectResults: []command.ProjectResult{
			{Failure: "This project is currently locked by an unapplied plan from pull #1."},
		},
	})
	require.NoError(t, err)
	return body
}

func TestDrifter_FindDriftedWorkspacesUsesRef(t *testing.T) {
	var mu sync.Mutex
	var refs []string
	body := lockedPlanResponse(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req controllers.APIRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("unable to decode plan request: %v", err)
		}
		mu.Lock()
		refs = append(refs, req.Ref)
		mu.Unlock()
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	logger := zaptest.NewLogger(t)
	d := Drifter{
		Logger: logger,
		Repo:   "owner/repo",
		AtlantisClient: &atlantis.Client{
			AtlantisHostname: srv.URL,
			Token:            "token",
			HTTPClient:       srv.Client(),
		},
		ResultCache:  processedcache.Noop{},
		Notification: &notification.Zap{Logger: logger},
	}
	ws := atlantis.DirectoriesWithWorkspaces{"dir": {"default"}}
	require.NoError(t, d.FindDriftedWorkspaces(context.Background(), "main", ws))
	require.Equal(t, []string{"main"}, refs)
}
