package jobs

import (
	"fmt"

	"github.com/spf13/viper"
	"gopkg.in/cyverse-de/model.v4"
)

// JobSubmissionBuilder is an interface for generating Condor job submissions.
type JobSubmissionBuilder interface {
	Build(submission *model.Job, dirPath string) (string, error)
}

// NewJobSubmissionBuilder returns an implementation of the JobSubmissionBuilder
// interface that's appropriate for the jobType that gets passed in. A job type
// of 'condor' is used for submissions to the local HTCondor cluster.
// 'interapps' is used for interactive app submissions.
func NewJobSubmissionBuilder(jobType string, cfg *viper.Viper) (JobSubmissionBuilder, error) {
	switch jobType {
	case "condor":
		return newCondorJobSubmissionBuilder(cfg), nil
	case "interapps":
		return InterappsSubmissionBuilder{cfg: cfg}, nil
	case "osg":
		return newOSGJobSubmissionBuilder(cfg), nil
	default:
		return nil, fmt.Errorf("unrecognized job submission type: %s", jobType)
	}
}
