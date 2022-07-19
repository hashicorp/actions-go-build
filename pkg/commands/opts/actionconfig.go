package opts

import "github.com/hashicorp/actions-go-build/internal/config"

// ActionConfig wraps a config.Config to add the ReadEnv method.
type ActionConfig struct {
	config.Config
}

func (ac *ActionConfig) ReadEnv() error {
	var err error
	ac.Config, err = config.FromEnvironment()
	return err
}
