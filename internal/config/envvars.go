// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package config

import (
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/sethvargo/go-githubactions"
)

type envSetter struct {
	setEnvFunc func(name, value string)
}

func newEnvSetter() envSetter {
	if os.Getenv("GITHUB_ENV") != "" {
		return envSetter{githubactions.SetEnv}
	}
	log.Printf("WARNING: GITHUB_ENV not set, just printing environment.")
	return envSetter{nil}
}

type EnvVar struct{ Name, Value string }

func (c Config) EnvVars() ([]EnvVar, error) {
	var kvs []EnvVar
	addEnv := func(key, value string) {
		kvs = append(kvs, EnvVar{key, value})
	}

	// TODO don't serialise primary and verification build configs to env here.
	// We can derive them from the rest of the config anyway so there's probably
	// no point writing them to GITHUB_ENV.
	//
	// Keeping them here for now since the current bash implementation expects
	// to see them.

	primary, err := c.PrimaryBuildConfig()
	if err != nil {
		return nil, err
	}
	verification, err := c.VerificationBuildConfig()
	if err != nil {
		return nil, err
	}

	addEnv("PRODUCT_REPOSITORY", c.Product.Repository)
	addEnv("PRODUCT_NAME", c.Product.Name)
	addEnv("PRODUCT_VERSION", c.Product.Version.Full)
	addEnv("PRODUCT_REVISION", c.Product.Revision)
	addEnv("PRODUCT_REVISION_TIME", c.Product.RevisionTime)
	addEnv("GO_VERSION", c.Parameters.GoVersion)
	addEnv("OS", c.Parameters.OS)
	addEnv("ARCH", c.Parameters.Arch)
	addEnv("REPRODUCIBLE", c.Reproducible)
	addEnv("INSTRUCTIONS", c.Parameters.Instructions)
	addEnv("BIN_NAME", c.Product.ExecutableName)
	addEnv("ZIP_NAME", c.Parameters.ZipName)
	addEnv("PRIMARY_BUILD_ROOT", c.Primary.BuildRoot)
	addEnv("VERIFICATION_BUILD_ROOT", c.Verification.BuildRoot)
	addEnv("PRIMARY_BUILD_RESULT", c.Primary.BuildResult)
	addEnv("VERIFICATION_BUILD_RESULT", c.Verification.BuildResult)
	addEnv("BIN_PATH_PRIMARY", primary.Paths.BinPath)
	addEnv("ZIP_PATH_PRIMARY", primary.Paths.ZipPath)
	addEnv("BIN_PATH_VERIFICATION", verification.Paths.BinPath)
	addEnv("ZIP_PATH_VERIFICATION", verification.Paths.ZipPath)
	addEnv("VERIFICATION_RESULT", c.VerificationResult)
	addEnv("DEBUG", strconv.FormatBool(c.Debug))

	return kvs, nil
}

func (c Config) foreach(fn func(key, value string)) error {
	vars, err := c.EnvVars()
	if err != nil {
		return err
	}
	for _, pair := range vars {
		fn(pair.Name, pair.Value)
	}
	return nil
}

// ExportToGitHubEnv writes this config to GITHUB_ENV so it can be read by
// future steps in this job. If GITHUB_ENV isn't set, it prints a warning
// and just logs what would have been set.
func (c Config) ExportToGitHubEnv() error {
	if os.Getenv("GITHUB_ENV") == "" {
		return errors.New("GITHUB_ENV not set")
	}
	es := newEnvSetter()
	return c.foreach(es.setEnv)
}

func (es envSetter) setEnv(name, value string) {
	log.Printf("Setting %q to %q", name, value)
	es.setEnvFunc(name, value)
}
