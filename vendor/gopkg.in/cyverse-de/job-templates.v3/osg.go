package jobs

import (
	"github.com/spf13/viper"
	"gopkg.in/cyverse-de/model.v2"
	"net/url"
	"path/filepath"
)

// OSGJobSubmissionBuilder is responsible for writing out the iplant.cmd, config.json,
// input_ticket.list, and output_ticket.list files for jobs that are sent to OSG.
type OSGJobSubmissionBuilder struct {
	cfg *viper.Viper
}

// OSGJobConfig stores configuration settings used by the tool wrapper script in the Singularity container.
type OSGJobConfig struct {
	IrodsHost        string `json:"irods_host"`
	IrodsPort        int    `json:"irods_port"`
	IrodsJobUser     string `json:"irods_job_user"`
	InputTicketList  string `json:"input_ticket_list"`
	OutputTicketList string `json:"output_ticket_list"`
	StatusUpdateUrl  string `json:"status_update_url"`
	Stdout           string `json:"stdout"`
	Stderr           string `json:"stderr"`
}

func (b OSGJobSubmissionBuilder) generateJobStatusUpdateUrl(submission *model.Job) string {
	return b.cfg.GetString("status_listener.url") +
		"/" + url.PathEscape(submission.InvocationID) +
		"/status"
}

// generateConfig builds the configuration structure without marshaling it. The primary reason this function
// exists is to make testing easier.
func (b OSGJobSubmissionBuilder) generateConfig(submission *model.Job) *OSGJobConfig {
	return &OSGJobConfig{
		IrodsHost:        b.cfg.GetString("external_irods.host"),
		IrodsPort:        b.cfg.GetInt("external_irods.port"),
		IrodsJobUser:     submission.Submitter,
		InputTicketList:  submission.InputTicketsFile,
		OutputTicketList: submission.OutputTicketFile,
		StatusUpdateUrl:  b.generateJobStatusUpdateUrl(submission),
		Stdout:           "out.txt",
		Stderr:           "err.txt",
	}
}

// generateConfigJson generates the config.json file which will be used by the wrapper script.
func (b OSGJobSubmissionBuilder) generateConfigJson(submission *model.Job, dirPath string) (string, error) {
	return generateJson(dirPath, "config.json", b.generateConfig(submission))
}

// Build is where the files are actually written out for submissions to OSG.
func (b OSGJobSubmissionBuilder) Build(submission *model.Job, dirPath string) (string, error) {
	var err error

	templateFields := OtherTemplateFields{TicketPathListHeader: b.cfg.GetString("tickets_path_list.file_identifier")}
	templateModel := TemplatesModel{
		submission,
		templateFields,
	}

	outputTicketFile, err := generateOutputTicketList(dirPath, templateModel)
	if err != nil {
		return "", err
	}
	submission.OutputTicketFile = filepath.Base(outputTicketFile)

	inputTicketFile, err := generateInputTicketList(dirPath, templateModel)
	if err != nil {
		return "", err
	}
	submission.InputTicketsFile = filepath.Base(inputTicketFile)

	configFile, err := b.generateConfigJson(submission, dirPath)
	if err != nil {
		return "", err
	}
	submission.ConfigFile = filepath.Base(configFile)

	// Generate the submission file.
	submitFilePath, err := generateFile(dirPath, "iplant.cmd", osgSubmissionTemplate, submission)
	if err != nil {
		return "", err
	}

	return submitFilePath, nil
}

func newOSGJobSubmissionBuilder(cfg *viper.Viper) OSGJobSubmissionBuilder {
	return OSGJobSubmissionBuilder{cfg: cfg}
}
