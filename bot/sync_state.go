package bot

import "github.com/google/go-github/v33/github"

type SyncState struct {
	Name          string
	ProjectColumn *github.ProjectColumn
	Labels        []*github.Label
}
