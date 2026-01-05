// Copyright IBM Corp. 2022, 2025
// SPDX-License-Identifier: MPL-2.0

package build

import (
	"os"
	"time"

	cp "github.com/otiai10/copy"
)

// LocalVerification is the local verification build. It is run inside a
// temporary copy of the primary build's root directory.
type LocalVerification struct {
	*core
	primaryRoot string
	startAfter  time.Time
}

func NewLocalVerification(primaryRoot string, startAfter time.Time, cfg Config, options ...Option) (Build, error) {
	options = append(options, AsVerificationBuild())
	core, err := newCore("local verification", cfg, options...)
	if err != nil {
		return nil, err
	}
	if err := core.ChangeToVerificationRoot(); err != nil {
		return nil, err
	}
	return &LocalVerification{
		core:        core,
		primaryRoot: primaryRoot,
		startAfter:  startAfter,
	}, nil
}

func (lv *LocalVerification) Kind() string { return "local verification" }

func (lv *LocalVerification) Steps() []Step {

	var sleepTime time.Duration
	now := time.Now()
	if lv.startAfter.After(now) {
		sleepTime = lv.startAfter.Sub(now)
	}
	pPath := lv.primaryRoot
	vPath := lv.Config().Paths.WorkDir

	pre := []Step{
		newStep("ensuring new empty directory to run build in", func() error {
			return os.RemoveAll(vPath)
		}),
		newStep("copying primary build root dir to temp dir", func() error {
			return cp.Copy(pPath, vPath)
		}),
		newStep("waiting until the stagger time has elapsed", func() error {
			time.Sleep(sleepTime)
			return nil
		}),
	}

	return append(pre, lv.core.Steps()...)
}
