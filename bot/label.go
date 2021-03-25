package bot

import (
	"context"
	"fmt"

	"github.com/google/go-github/v33/github"
)

func (b *Bot) getLabel(name string) (*github.Label, error) {
	label, _, err := b.gh.Issues.GetLabel(context.Background(), b.conf.Owner, b.conf.Repository, name)
	if err != nil {
		return nil, fmt.Errorf("No label named '%s' exist in repository\n", name)
	}
	return label, nil
}
