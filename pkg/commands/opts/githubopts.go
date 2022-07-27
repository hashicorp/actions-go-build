package opts

import (
	"flag"
	"os"
)

type GitHubOpts struct {
	GitHubMode bool
}

func (gho *GitHubOpts) ReadEnv() error {
	gho.GitHubMode = os.Getenv("GITHUB_ACTIONS") == "true"
	return nil
}

func (gho *GitHubOpts) Flags(fs *flag.FlagSet) {
	fs.BoolVar(&gho.GitHubMode, "github", gho.GitHubMode, "run as though on GitHub Actions")
}
