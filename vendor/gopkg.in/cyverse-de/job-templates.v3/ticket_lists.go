package jobs

import (
	"bytes"
	"path/filepath"
)

func generateOutputTicketList(dirPath string, submission TemplatesModel) (string, error) {
	if submission.OutputDirTicket != "" {
		// Generate the output ticket path list file.
		filePath, err := generateFile(dirPath, "output_ticket.list", outputTicketListTemplate, submission)
		return filepath.Base(filePath), err
	}

	return "", nil
}

func generateOutputTicketListContents(submission TemplatesModel) (*bytes.Buffer, error) {
	if submission.OutputDirTicket != "" {
		return generateFileContents(outputTicketListTemplate, submission)
	}

	return nil, nil
}

func generateInputTicketList(dirPath string, submission TemplatesModel) (string, error) {
	if len(submission.FilterInputsWithTickets()) > 0 {
		// Generate the input tickets path list file.
		filePath, err := generateFile(dirPath, "input_ticket.list", inputTicketListTemplate, submission)
		return filepath.Base(filePath), err
	}

	return "", nil
}

func generateInputTicketListContents(submission TemplatesModel) (*bytes.Buffer, error) {
	if len(submission.FilterInputsWithTickets()) > 0 {
		return generateFileContents(inputTicketListTemplate, submission)
	}

	return nil, nil
}
