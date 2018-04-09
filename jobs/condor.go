package jobs

import (
	"github.com/spf13/viper"
	"gopkg.in/cyverse-de/model.v2"
)

type CondorJobSubmissionBuilder struct {
	cfg *viper.Viper
}

func (b CondorJobSubmissionBuilder) Build(submission *model.Job, dirPath string) (string, error) {
	var err error

	templateFields := OtherTemplateFields{TicketPathListHeader: b.cfg.GetString("tickets_path_list.file_identifier")}
	templateModel := TemplatesModel{
		submission,
		templateFields,
	}

	submission.OutputTicketFile, err = generateOutputTicketList(dirPath, templateModel)
	if err != nil {
		return "", err
	}

	submission.InputTicketsFile, err = generateInputTicketList(dirPath, templateModel)
	if err != nil {
		return "", err
	}

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
func generateOutputTicketList(dirPath string, submission TemplatesModel) (string, error) {
	if submission.OutputDirTicket != "" {
		// Generate the output ticket path list file.
		return generateFile(dirPath, "output_ticket.list", outputTicketListTemplate, submission)
	}

	return "", nil
}

func generateInputTicketList(dirPath string, submission TemplatesModel) (string, error) {
	if len(submission.FilterInputsWithTickets()) > 0 {
		// Generate the input tickets path list file.
		return generateFile(dirPath, "input_ticket.list", inputTicketListTemplate, submission)
	}

	return "", nil
}

func newCondorJobSubmissionBuilder(cfg *viper.Viper) JobSubmissionBuilder {
	return CondorJobSubmissionBuilder{cfg: cfg}
}
