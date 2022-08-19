package atlantis

import (
	"context"
	"encoding/json"
	"github.com/cresta/atlantis-drift-detection/internal/testhelper"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"testing"
)

func makeTestClient(t *testing.T) *Client {
	atlantisHost := os.Getenv("ATLANTIS_HOST")
	if atlantisHost == "" {
		t.Skip("ATLANTIS_HOST not set, skipping test")
	}
	testToken := os.Getenv("ATLANTIS_TOKEN")
	if testToken == "" {
		t.Skip("ATLANTIS_TOKEN not set, skipping test")
	}
	c := Client{
		AtlantisHostname: atlantisHost,
		Token:            testToken,
		HTTPClient:       http.DefaultClient,
	}
	return &c
}

func loadPlanSummaryOk(t *testing.T) *PlanSummaryRequest {
	body := os.Getenv("PLAN_SUMMARY_OK")
	if body == "" {
		t.Skip("PLAN_SUMMARY_OK not set, skipping test")
	}
	var ret PlanSummaryRequest
	require.NoError(t, json.Unmarshal([]byte(body), &ret))
	return &ret
}

func loadPlanSummaryChanges(t *testing.T) *PlanSummaryRequest {
	body := os.Getenv("PLAN_SUMMARY_CHANGES")
	if body == "" {
		t.Skip("PLAN_SUMMARY_CHANGES not set, skipping test")
	}
	var ret PlanSummaryRequest
	require.NoError(t, json.Unmarshal([]byte(body), &ret))
	return &ret
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
