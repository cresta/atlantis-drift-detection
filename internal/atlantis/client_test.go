package atlantis

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cresta/atlantis-drift-detection/internal/testhelper"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeFakeAtlantis(t *testing.T, code int, body string) *Client {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/plan" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(code)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return &Client{
		AtlantisHostname: srv.URL,
		Token:            "token",
		HTTPClient:       srv.Client(),
	}
}

func fakePlanSummaryRequest() *PlanSummaryRequest {
	return &PlanSummaryRequest{
		Repo:      "owner/repo",
		Ref:       "main",
		Type:      "Github",
		Dir:       "dir",
		Workspace: "default",
	}
}

func TestClient_PlanSummaryServerError(t *testing.T) {
	c := makeFakeAtlantis(t, http.StatusInternalServerError, `{"error":"running git fetch origin pull/0/head:: fatal: couldn't find remote ref pull/0/head"}`+"\n")
	_, err := c.PlanSummary(context.Background(), fakePlanSummaryRequest())
	require.Error(t, err)
	require.Contains(t, err.Error(), "couldn't find remote ref pull/0/head")
	var tmp TemporaryError
	require.ErrorAs(t, err, &tmp)
	require.True(t, tmp.Temporary())
}

func TestClient_PlanSummaryBadRequest(t *testing.T) {
	c := makeFakeAtlantis(t, http.StatusBadRequest, `{"error":"request is missing fields"}`+"\n")
	_, err := c.PlanSummary(context.Background(), fakePlanSummaryRequest())
	require.Error(t, err)
	require.Contains(t, err.Error(), "request is missing fields")
	var tmp TemporaryError
	require.False(t, errors.As(err, &tmp))
}

func TestClient_PlanSummaryLockedResult(t *testing.T) {
	body, err := json.Marshal(command.Result{
		ProjectResults: []command.ProjectResult{
			{Failure: "This project is currently locked by an unapplied plan from pull #1."},
		},
	})
	require.NoError(t, err)
	c := makeFakeAtlantis(t, http.StatusInternalServerError, string(body)+"\n")
	res, err := c.PlanSummary(context.Background(), fakePlanSummaryRequest())
	require.NoError(t, err)
	require.True(t, res.IsLocked())
}

func TestClient_PlanSummaryUndecodableBody(t *testing.T) {
	c := makeFakeAtlantis(t, http.StatusServiceUnavailable, "upstream connect error\n")
	_, err := c.PlanSummary(context.Background(), fakePlanSummaryRequest())
	require.Error(t, err)
	require.Contains(t, err.Error(), "(body:upstream connect error")
}

func makeTestClient(t *testing.T) *Client {
	c := Client{
		AtlantisHostname: testhelper.EnvOrSkip(t, "ATLANTIS_HOST"),
		Token:            testhelper.EnvOrSkip(t, "ATLANTIS_TOKEN"),
		HTTPClient:       http.DefaultClient,
	}
	return &c
}

func loadPlanSummaryOk(t *testing.T) *PlanSummaryRequest {
	return loadPlanSummaryRequest(t, "PLAN_SUMMARY_OK")
}

func loadPlanSummaryRequest(t *testing.T, name string) *PlanSummaryRequest {
	body := testhelper.EnvOrSkip(t, name)
	var ret PlanSummaryRequest
	require.NoError(t, json.Unmarshal([]byte(body), &ret))
	return &ret
}

func loadPlanSummaryLock(t *testing.T) *PlanSummaryRequest {
	return loadPlanSummaryRequest(t, "PLAN_SUMMARY_LOCK")
}

func loadPlanSummaryChanges(t *testing.T) *PlanSummaryRequest {
	return loadPlanSummaryRequest(t, "PLAN_SUMMARY_CHANGES")
}

func TestClient_PlanSummaryLock(t *testing.T) {
	testhelper.ReadEnvFile(t, "../../")
	c := makeTestClient(t)
	req := loadPlanSummaryLock(t)
	ok, err := c.PlanSummary(context.Background(), req)
	require.NoError(t, err)
	require.True(t, ok.IsLocked())
}

func TestClient_PlanSummaryOk(t *testing.T) {
	testhelper.ReadEnvFile(t, "../../")
	c := makeTestClient(t)
	req := loadPlanSummaryOk(t)
	ok, err := c.PlanSummary(context.Background(), req)
	require.NoError(t, err)
	require.False(t, ok.HasChanges())
}

func TestClient_PlanSummaryChanges(t *testing.T) {
	testhelper.ReadEnvFile(t, "../../")
	c := makeTestClient(t)
	req := loadPlanSummaryChanges(t)
	ok, err := c.PlanSummary(context.Background(), req)
	require.NoError(t, err)
	require.True(t, ok.HasChanges())
}
