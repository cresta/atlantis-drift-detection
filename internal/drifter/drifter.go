package drifter

import (
	"context"
	"fmt"
	"github.com/cresta/gogit"
	"os"
)

type Drifter struct {
	Repo   string
	Cloner gogit.Cloner
}

func (d *Drifter) Drift(ctx context.Context) error {
	repository, err := d.Cloner.Clone(ctx, d.Repo)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(repository.Location()); err != nil {
			panic(err)
		}
	}()
	return nil
}
