package jobs

import "bytes"

func generateOutputTicketList(dirPath string, submission TemplatesModel) (string, error) {
	if submission.OutputDirTicket != "" {
		// Generate the output ticket path list file.
		return generateFile(dirPath, "output_ticket.list", outputTicketListTemplate, submission)
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
		return generateFile(dirPath, "input_ticket.list", inputTicketListTemplate, submission)
	}

	return "", nil
}

func generateInputTicketListContents(submission TemplatesModel) (*bytes.Buffer, error) {
	if len(submission.FilterInputsWithTickets()) > 0 {
		return generateFileContents(inputTicketListTemplate, submission)
	}

	return nil, nil
}
