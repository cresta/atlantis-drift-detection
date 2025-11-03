package atlantisgithub

import (
	"context"
	"fmt"
	"github.com/cresta/gogit"
	"github.com/cresta/gogithub"
)

func CheckOutTerraformRepo(ctx context.Context, gitHubClient gogithub.GitHub, cloner *gogit.Cloner, repo string, repoRef string) (*gogit.Repository, error) {
	token, err := gitHubClient.GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	// https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps#http-based-git-access-by-an-installation
	githubRepoURL := fmt.Sprintf("https://x-access-token:%s@github.com/%s.git", token, repo)
	repository, err := cloner.Clone(ctx, githubRepoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repo: %w", err)
	}

	// Check if we're already on the desired branch (common case when ref is the default branch)
	currentBranch, err := repository.CurrentBranchName(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	// Only checkout if we're not already on the desired branch
	if currentBranch != repoRef {
		if err := repository.CheckoutNewBranch(ctx, repoRef); err != nil {
			return nil, fmt.Errorf("failed to checkout branch %s: %w", repoRef, err)
		}
	}

	return repository, nil
}
