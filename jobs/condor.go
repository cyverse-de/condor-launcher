package jobs

import (
	"github.com/spf13/viper"
	"gopkg.in/cyverse-de/model.v1"
)

type CondorJobSubmissionBuilder struct {
	cfg *viper.Viper
}

func (b CondorJobSubmissionBuilder) build(submission *model.Job, dirPath string) (*string, error) {

	// Generate the submission file.
	submitFilePath, err := generateFile(dirPath, "iplant.cmd", condorSubmissionTemplate, submission)
	if err != nil {
		return nil, err
	}

	// Generate the job configuration file.
	_, err = generateFile(dirPath, "config", condorJobConfigTemplate, b.cfg)
	if err != nil {
		return nil, err
	}

	// Write the job submission to a JSON file.
	_, err = generateJson(dirPath, "job", submission)
	if err != nil {
		return nil, err
	}

	return submitFilePath, nil
}

func newCondorJobSubmissionBuilder(cfg *viper.Viper) JobSubmissionBuilder {
	return CondorJobSubmissionBuilder{cfg: cfg}
}
