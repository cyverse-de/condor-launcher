package jobs

import (
	"github.com/spf13/viper"
	"gopkg.in/cyverse-de/model.v2"
	"path/filepath"
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
	templateFields := OtherTemplateFields{
		PathListHeader: b.cfg.GetString("path_list.file_identifier"),
		TicketPathListHeader: b.cfg.GetString("tickets_path_list.file_identifier"),
	}
	templateModel := TemplatesModel{
		submission,
		templateFields,
	}

	outputTicketFile, err := generateOutputTicketList(dirPath, templateModel)
	if err != nil {
		return "", err
	}
	submission.OutputTicketFile = filepath.Base(outputTicketFile)

	inputTicketsFile, err := generateInputTicketList(dirPath, templateModel)
	if err != nil {
		return "", err
	}
	submission.InputTicketsFile = filepath.Base(inputTicketsFile)

	inputPathListFile, err := generateInputPathList(dirPath, templateModel)
	if err != nil {
		return "", err
	}
	submission.InputPathListFile = filepath.Base(inputPathListFile)

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
	_, err = generateJSON(dirPath, "job", submission)
	if err != nil {
		return "", err
	}

	return submitFilePath, nil
}

func generateInputPathList(dirPath string, submission TemplatesModel) (string, error) {
	if len(submission.FilterInputsWithoutTickets()) > 0 {
		// Generate the input path list file.
		return generateFile(dirPath, "input_path.list", inputPathListTemplate, submission)
	}

	return "", nil
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
	_, err = generateJSON(dirPath, "job", submission)
	if err != nil {
		return "", err
	}

	return submitFilePath, nil
}
