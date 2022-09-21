package commands

import (
	"flag"
	"fmt"
	"time"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/composite-action-framework-go/pkg/json"
)

// A verifyish represents something that can be verified as reproducible.
// To verify reproducibility, you need to compare two build results to see if they have
// the same SHAs for the zip files. Build results can come from files on disk, or
// URLs read from the internet, or from builds run locally on this machine.
//
// Verifyish takes a single argument "target" (the one defined in buildish), which represents
// the build we are wanting to verify (the "primary" build). That can be a build result file,
// or a file containing build config we can run, or a path to a local directory containing
// source code we can build. If a build needs to be run to obtain a build result then that
// is done first.
//
// Using this primary build result, a verification build is configured and run, and the results
// are compared.
//
// It is possible to skip the verification build by passing the -verification-build-result flag
// which allows you to directly compare a primary and verification build result which have
// been generated earlier. This is mostly useful in CI where you want to be able to generate
// multiple results in different jobs and then compare them in another job.
type verifyish struct {
	buildish
	staggerTime                 time.Duration
	verificationBuildResultFile string

	primary      build.ResultSource
	verification build.ResultSource
}

func (v *verifyish) Flags(fs *flag.FlagSet) {
	v.buildish.Flags(fs)
	fs.DurationVar(&v.staggerTime, "staggertime", 5*time.Second, "minimum time to wait after start of primary build")
	fs.StringVar(&v.verificationBuildResultFile, "verification-build-result", "", "load verification build result from file")
}

func (v *verifyish) Init() error {
	if err := v.buildish.Init(); err != nil {
		return err
	}
	return v.setResultSources()
}

func (v *verifyish) runVerification() (*build.VerificationResult, error) {
	verifier, err := v.buildish.buildFlags.newVerifier(v.primary, v.verification)
	if err != nil {
		return nil, err
	}
	return verifier.Verify()
}

func (v *verifyish) setResultSources() error {
	var err error
	if v.primary, err = v.primaryResultSource(); err != nil {
		return err
	}

	if v.verification, err = v.verificationResultSource(); err != nil {
		return err
	}

	return nil
}

func (v *verifyish) primaryResultSource() (build.ResultSource, error) {
	// Check if we were handed a result already (i.e. if the input was a build result or verification result file).
	if v.buildish.buildResult != nil {
		return v.buildish.buildResult, nil
	}
	return v.buildish.build("Getting primary build result", build.WithLogPrefix("primary build"))
}

func (v *verifyish) verificationResultSource() (build.ResultSource, error) {
	if v.verificationBuildResultFile != "" {
		// The user supplied a ready-made verification build result.
		v.log("Getting verification build result from %q", v.verificationBuildResultFile)
		return v.verificationResultSourceFromFile(v.verificationBuildResultFile)
	}
	// No ready-made verification build result, so we need to run a new one.
	v.log("Running new verification build.")
	return v.verificationResultSourceFromNewBuild()
}

func (v *verifyish) verificationResultSourceFromFile(path string) (build.ResultSource, error) {
	return json.ReadFile[build.Result](path)
}

func (v *verifyish) verificationResultSourceFromNewBuild() (build.ResultSource, error) {
	config, err := v.primaryConfig()
	if err != nil {
		return nil, err
	}

	opts := []build.Option{
		build.WithLogPrefix("verification build"),
		build.AsVerificationBuild(),
	}

	// If the primary build is sourced from a dir, we have the source on-disk.
	// This makes it a local verification build.
	if v.buildish.dir != "" {
		startAfter, err := v.calculateEarliestBuildTime()
		if err != nil {
			return nil, err
		}
		return v.buildish.buildFlags.newLocalVerificationManager(v.buildish.dir, startAfter, *config, opts...)
	}
	return v.buildish.buildFlags.newRemoteVerificationManager(*config, opts...)
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
	if v.buildish.storedBuild != nil {
		c := v.buildish.storedBuild.Config()
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

	pb, err := v.buildish.build("Inspecting cache for build defined")
	if err != nil {
		return build.Result{}, false, err
	}
	return pb.Build().CachedResult()
}
