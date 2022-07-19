package opts

import "github.com/hashicorp/actions-go-build/pkg/build"

type AllBuilds struct {
	Primary, Verification build.Build
}

func (ab *AllBuilds) ReadEnv() error {
	pb, vb := &PrimaryBuild{}, &VerificationBuild{}
	if err := pb.ReadEnv(); err != nil {
		return err
	}
	if err := vb.ReadEnv(); err != nil {
		return err
	}
	ab.Primary, ab.Verification = pb.Build, vb.Build
	return nil
}
