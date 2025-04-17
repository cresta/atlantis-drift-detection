package atlantis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/events/command"
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
	Summaries []PlanSummary
}

type PlanSummary struct {
	HasLock bool
	Summary string
}

func (p *PlanResult) HasChanges() bool {
	for _, summary := range p.Summaries {
		if summary.HasLock {
			continue
		}
		if !strings.Contains(summary.Summary, "No changes. ") {
			return true
		}
	}
	return false
}

func (p *PlanResult) IsLocked() bool {
	for _, summary := range p.Summaries {
		if !summary.HasLock {
			return false
		}
	}
	return true
}

type possiblyTemporaryError struct {
	error
}

type TemporaryError interface {
	Temporary() bool
	error
}

// Updated errorResponse to handle dynamic error types
type errorResponse struct {
	Error interface{} `json:"error"` // Allow both string and structured errors
}

func (p *possiblyTemporaryError) Temporary() bool {
	return true
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
	var fullBody bytes.Buffer
	if _, err := io.Copy(&fullBody, resp.Body); err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}
	if err := resp.Body.Close(); err != nil {
		return nil, fmt.Errorf("unable to close response body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		var errResp errorResponse
		if err := json.NewDecoder(&fullBody).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("unauthorized request to %s: %w", destination, err)
		}

		// Handle different error types dynamically
		switch errResp.Error := errResp.Error.(type) {
		case string:
			// Simple string error
			return nil, fmt.Errorf("unauthorized request to %s: %s", destination, errResp.Error)
		case map[string]interface{}:
			// Structured error object
			return nil, fmt.Errorf("unauthorized request to %s: %v", destination, errResp.Error)
		default:
			// Unknown error format
			return nil, fmt.Errorf("unknown error format in response: %v", errResp.Error)
		}
	}

	var bodyResult command.Result
	if err := json.NewDecoder(&fullBody).Decode(&bodyResult); err != nil {
		retErr := fmt.Errorf("error decoding plan response(code:%d)(status:%s)(body:%s): %w", resp.StatusCode, resp.Status, fullBody.String(), err)
		if resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusInternalServerError {
			// Handle temporary errors from Atlantis
			return nil, &possiblyTemporaryError{retErr}
		}
		return nil, retErr
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusInternalServerError {
		return nil, fmt.Errorf("non-200 and non-500 response for %s: %d", destination, resp.StatusCode)
	}

	// Check for errors in bodyResult
	if bodyResult.Error != nil {
		return nil, fmt.Errorf("error making plan request: %w", bodyResult.Error)
	}
	if bodyResult.Failure != "" {
		return nil, fmt.Errorf("failure making plan request: %s", bodyResult.Failure)
	}

	// Process project results
	var ret PlanResult
	for _, result := range bodyResult.ProjectResults {
		if result.Failure != "" {
			if strings.Contains(result.Failure, "This project is currently locked ") {
				ret.Summaries = append(ret.Summaries, PlanSummary{HasLock: true})
				continue
			}
		}
		if result.PlanSuccess != nil {
			summary := result.PlanSuccess.Summary()
			ret.Summaries = append(ret.Summaries, PlanSummary{Summary: summary})
			continue
		}
		return nil, fmt.Errorf("project result unknown failure: %s", result.Failure)
	}

	return &ret, nil
}
