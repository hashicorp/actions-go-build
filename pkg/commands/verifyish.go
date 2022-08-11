package commands

import (
	"flag"
	"fmt"
	"time"

	"github.com/hashicorp/actions-go-build/pkg/build"
)

type verifyish struct {
	buildish
	staggerTime time.Duration

	primary      build.ResultSource
	verification build.ResultSource
}

func (v *verifyish) Flags(fs *flag.FlagSet) {
	v.buildish.Flags(fs)
	fs.DurationVar(&v.staggerTime, "staggertime", 5*time.Second, "min. time to wait after start of primary build")
}

func (v *verifyish) Init() error {
	if err := v.buildish.Init(); err != nil {
		return err
	}
	return v.setResultSources()
}

func (v *verifyish) runVerification() error {
	verifier, err := v.buildish.buildFlags.newVerifier(v.primary, v.verification)
	if err != nil {
		return err
	}
	result, err := verifier.Verify()
	if err != nil {
		return err
	}
	return v.output.result("Reproducibility verification", result)
}

func (v *verifyish) setResultSources() error {
	var err error
	if v.primary, err = v.primaryResultSource(); err != nil {
		return err
	}

	primaryConfig, err := v.primaryConfig()
	if err != nil {
		return err
	}

	verificationConfig, err := primaryConfig.ChangeToVerificationRoot()
	if err != nil {
		return err
	}

	// If the primary build is sourced from a dir, we have the source on-disk.
	// This makes it a local verification build.
	if v.buildish.dir != "" {
		startAfter, err := v.calculateEarliestBuildTime()
		if err != nil {
			return err
		}
		return v.configureLocalVerificationBuild(v.buildish.dir, startAfter, verificationConfig)
	}

	return v.configureRemoteVerificationBuild(verificationConfig)
}

func (v *verifyish) primaryResultSource() (build.ResultSource, error) {
	// Check if we were handed a result already (i.e. if the input was a build or verification result file).
	if v.buildish.buildResult != nil {
		return v.buildish.buildResult, nil
	}
	return v.buildish.Build("Getting primary build result", false, build.WithLogPrefix("primary build"))
}

func (v *verifyish) configureLocalVerificationBuild(dir string, startAfter time.Time, c build.Config) error {
	logPrefix := build.WithLogPrefix("verification build")
	//	b, err := v.buildish.Build("Getting local verification build result", true, logPrefix)
	//	if err != nil {
	//		return err
	//	}
	//	if err := b.Build().ChangeToVerificationRoot(); err != nil {
	//		return err
	//	}
	var err error
	v.verification, err = v.buildish.buildFlags.newLocalVerificationManager(v.buildish.dir, startAfter, c, logPrefix)
	return err
}

func (v *verifyish) configureRemoteVerificationBuild(c build.Config) error {
	var err error
	v.verification, err = v.buildish.buildFlags.newRemoteVerificationManager(c, build.WithLogPrefix("verification build"))
	return err
}

func (v *verifyish) calculateEarliestBuildTime() (time.Time, error) {
	// By default, we'll just wait the staggerTime.
	startAfter := time.Now().Add(v.staggerTime)
	primaryResult, ready, err := v.readyPrimaryResult()
	if err != nil || !ready {
		return startAfter, err
	}
	// If we know when the primary build started we can go a bit quicker.
	return primaryResult.Meta.Start.Add(v.staggerTime), nil
}

func (v *verifyish) primaryConfig() (*build.Config, error) {
	if v.buildish.buildConfig != nil {
		return v.buildish.buildConfig, nil
	}
	if v.buildish.build != nil {
		c := v.buildish.build.Config()
		return &c, nil
	}
	return nil, fmt.Errorf("no build config to compare")
}

// readyPrimaryResult returns the primary build result (the one we're comparing
// only if it's already available, i.e. the build has been run or we're loading
// the result from a file).
func (v *verifyish) readyPrimaryResult() (build.Result, bool, error) {
	if v.buildish.buildResult != nil {
		return *v.buildish.buildResult, true, nil
	}
	if v.buildish.dir == "" {
		return build.Result{}, false, nil
	}

	pb, err := v.buildish.Build("Inspecting cache for build defined", false)
	if err != nil {
		return build.Result{}, false, err
	}
	return pb.Build().CachedResult()
}
