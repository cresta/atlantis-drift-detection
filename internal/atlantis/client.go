package atlantis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/events/command"
	"net/http"
	"strings"
)

type Client struct {
	AtlantisHostname string
	Token            string
	HTTPClient       *http.Client
}

type PlanSummaryRequest struct {
	Repo      string
	Ref       string
	Type      string
	Dir       string
	Workspace string
}

type PlanResult struct {
	Summaries []string
}

func (p *PlanResult) HasChanges() bool {
	for _, summary := range p.Summaries {
		if !strings.Contains(summary, "No changes. ") {
			return true
		}
	}
	return false
}

func (c *Client) PlanSummary(ctx context.Context, req *PlanSummaryRequest) (*PlanResult, error) {
	planBody := controllers.APIRequest{
		Repository: req.Repo,
		Ref:        req.Ref,
		Type:       req.Type,
		Paths: []struct {
			Directory string
			Workspace string
		}{
			{
				Directory: req.Dir,
				Workspace: req.Workspace,
			},
		},
	}
	planBodyJSON, err := json.Marshal(planBody)
	if err != nil {
		return nil, fmt.Errorf("error marshalling plan body: %w", err)
	}
	destination := fmt.Sprintf("%s/api/plan", c.AtlantisHostname)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, destination, strings.NewReader(string(planBodyJSON)))
	if err != nil {
		return nil, fmt.Errorf("error parsing destination: %w", err)
	}
	httpReq.Header.Set("X-Atlantis-Token", c.Token)
	httpReq = httpReq.WithContext(ctx)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error making plan request to %s: %w", destination, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 response for %s: %d", destination, resp.StatusCode)
	}
	var bodyResult command.Result
	if err := json.NewDecoder(resp.Body).Decode(&bodyResult); err != nil {
		return nil, fmt.Errorf("error decoding plan response: %s", err)
	}
	if bodyResult.Error != nil {
		return nil, fmt.Errorf("error making plan request: %w", bodyResult.Error)
	}
	if bodyResult.Failure != "" {
		return nil, fmt.Errorf("failure making plan request: %s", bodyResult.Failure)
	}
	var ret PlanResult
	for _, result := range bodyResult.ProjectResults {
		if result.Error != nil {
			return nil, fmt.Errorf("error getting project result: %w", result.Error)
		}
		if !result.IsSuccessful() {
			return nil, fmt.Errorf("plan result not successful: %s", result.Failure)
		}
		if result.Failure != "" {
			return nil, fmt.Errorf("project result failure: %s", result.Failure)
		}
		if result.PlanSuccess == nil {
			return nil, fmt.Errorf("plan result missing plan success")
		}
		summary := result.PlanSuccess.Summary()
		ret.Summaries = append(ret.Summaries, summary)
	}
	return &ret, nil
}
