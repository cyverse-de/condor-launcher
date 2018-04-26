package jobs

import (
	"github.com/spf13/viper"
	"gopkg.in/cyverse-de/model.v2"
)

// CondorJobSubmissionBuilder is responsible for writing out the iplant.cmd,
// config, and job files in the directory specififed by dirPath, but only for
// job submissions to our local HTCondor cluster.
type CondorJobSubmissionBuilder struct {
	cfg *viper.Viper
}

// Build is where the the iplant.cmd, config, and job files are actually written
// out for submissions to the local HTCondor cluster.
func (b CondorJobSubmissionBuilder) Build(submission *model.Job, dirPath string) (string, error) {

	// Generate the submission file.
	submitFilePath, err := generateFile(dirPath, "iplant.cmd", condorSubmissionTemplate, submission)
	if err != nil {
		return "", err
	}

	// Generate the job configuration file.
	_, err = generateFile(dirPath, "config", condorJobConfigTemplate, b.cfg)
	if err != nil {
		return "", err
	}

	// Write the job submission to a JSON file.
	_, err = generateJson(dirPath, "job", submission)
	if err != nil {
		return "", err
	}

	return submitFilePath, nil
}

func newCondorJobSubmissionBuilder(cfg *viper.Viper) JobSubmissionBuilder {
	return CondorJobSubmissionBuilder{cfg: cfg}
}

// InterappsSubmissionBuilder constructs the iplant.cmd, config, and job files
// in the directory indicated by dirPath, but only for interactive app job
// submissions.
type InterappsSubmissionBuilder struct {
	cfg *viper.Viper
}

// Build is where the iplant.cmd, config, and job files are actually written
// out. Satisfies the JobSubmissionBuilder interface.
func (b InterappsSubmissionBuilder) Build(submission *model.Job, dirPath string) (string, error) {
	// Generate the submission file.
	submitFilePath, err := generateFile(dirPath, "iplant.cmd", interappsSubmissionTemplate, submission)
	if err != nil {
		return "", err
	}

	// Generate the job configuration file.
	_, err = generateFile(dirPath, "config", interappsJobConfigTemplate, b.cfg)
	if err != nil {
		return "", err
	}

	// Write the job submission to a JSON file.
	_, err = generateJson(dirPath, "job", submission)
	if err != nil {
		return "", err
	}

	return submitFilePath, nil
}
