package jobs

import (
	"github.com/spf13/viper"
	"gopkg.in/cyverse-de/model.v1"
	"fmt"
)

// JobSubmissionBuilder is an interface for generating Condor job submissions.
type JobSubmissionBuilder interface {
	build(submission *model.Job, dirPath string) (*string, error)
}

func NewJobSubmissionBuilder(jobType string, cfg *viper.Viper) (JobSubmissionBuilder, error) {
	if jobType == "condor" {
		return newCondorJobSubmissionBuilder(cfg), nil
	}
	return nil, fmt.Errorf("unrecognized job submission type: %s\n", jobType)
}
