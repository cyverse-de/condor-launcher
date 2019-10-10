package jobs

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/cyverse-de/model.v4"
)

// OSGJobSubmissionBuilder is responsible for writing out the iplant.cmd, config.json,
// input_ticket.list, and output_ticket.list files for jobs that are sent to OSG.
type OSGJobSubmissionBuilder struct {
	cfg *viper.Viper
}

// OSGJobConfig stores configuration settings used by the tool wrapper script in the Singularity container.
type OSGJobConfig struct {
	Arguments        []string `json:"arguments"`
	IrodsHost        string   `json:"irods_host"`
	IrodsPort        int      `json:"irods_port"`
	IrodsJobUser     string   `json:"irods_job_user"`
	IrodsUsername    string   `json:"irods_user_name"`
	IrodsZoneName    string   `json:"irods_zone_name"`
	InputTicketList  string   `json:"input_ticket_list"`
	OutputTicketList string   `json:"output_ticket_list"`
	StatusUpdateURL  string   `json:"status_update_url"`
	Stdout           string   `json:"stdout"`
	Stderr           string   `json:"stderr"`
}

// generateJobStatusUpdateURL builds the URL that the job can use to notify the user when
// the job status has changed.
func (b OSGJobSubmissionBuilder) generateJobStatusUpdateURL(submission *model.Job) string {
	return b.cfg.GetString("status_listener.url") +
		"/" + url.PathEscape(submission.InvocationID) +
		"/status"
}

// stepArguments builds the list of arguments to pass to a job, returning an empty list of there
// are no arguments to pass to the job.
func (b OSGJobSubmissionBuilder) stepArguments(step model.Step) []string {
	arguments := step.Arguments()
	if arguments == nil {
		return []string{}
	}
	return arguments
}

// generateConfig builds the configuration structure without marshaling it. The primary reason this function
// exists is to make testing easier.
func (b OSGJobSubmissionBuilder) generateConfig(submission *model.Job) *OSGJobConfig {
	return &OSGJobConfig{
		Arguments:        b.stepArguments(submission.Steps[0]),
		IrodsHost:        b.cfg.GetString("external_irods.host"),
		IrodsPort:        b.cfg.GetInt("external_irods.port"),
		IrodsJobUser:     submission.Submitter,
		IrodsUsername:    b.cfg.GetString("external_irods.user"),
		IrodsZoneName:    "",
		InputTicketList:  submission.InputTicketsFile,
		OutputTicketList: submission.OutputTicketFile,
		StatusUpdateURL:  b.generateJobStatusUpdateURL(submission),
		Stdout:           "out.txt",
		Stderr:           "err.txt",
	}
}

// generateConfigJSON generates the config.json file which will be used by the wrapper script.
func (b OSGJobSubmissionBuilder) generateConfigJSON(submission *model.Job, dirPath string) (string, error) {
	return generateJSON(dirPath, "config.json", b.generateConfig(submission))
}

// Build is where the files are actually written out for submissions to OSG.
func (b OSGJobSubmissionBuilder) Build(submission *model.Job, dirPath string) (string, error) {
	var err error

	// Validate the submission.
	if len(submission.Steps) != 1 {
		return "", fmt.Errorf("only single-step OSG jobs are supported at this time")
	}

	// Build the template model for the input and output ticket lists.
	templateFields := OtherTemplateFields{TicketPathListHeader: b.cfg.GetString("tickets_path_list.file_identifier")}
	templateModel := TemplatesModel{
		submission,
		templateFields,
	}

	// Generate the list of output tickets.
	submission.OutputTicketFile, err = generateOutputTicketList(dirPath, templateModel)
	if err != nil {
		return "", err
	}

	// Generate the list of input tickets.
	submission.InputTicketsFile, err = generateInputTicketList(dirPath, templateModel)
	if err != nil {
		return "", err
	}

	// Generate the job configuration file.
	configFile, err := b.generateConfigJSON(submission, dirPath)
	if err != nil {
		return "", err
	}
	submission.ConfigFile = filepath.Base(configFile)

	// Generate the submission file.
	submitFilePath, err := generateFile(dirPath, "iplant.cmd", osgSubmissionTemplate, submission)
	if err != nil {
		return "", err
	}

	// Generate the job JSON file for debugging purposes.
	_, err = generateJSON(dirPath, "job", submission)
	if err != nil {
		return "", err
	}

	return submitFilePath, nil
}

func newOSGJobSubmissionBuilder(cfg *viper.Viper) OSGJobSubmissionBuilder {
	return OSGJobSubmissionBuilder{cfg: cfg}
}
