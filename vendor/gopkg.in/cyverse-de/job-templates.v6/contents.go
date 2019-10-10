package jobs

import (
	"bytes"
	"fmt"

	"gopkg.in/cyverse-de/model.v4"
)

// ExcludesFileContents returns a *bytes.Buffer containing the contents of an
// file exclusion list that gets passed to porklock to prevent it from uploading
// content. It's possible that the buffer is empty, but it shouldn't be nil.
func ExcludesFileContents(job *model.Job) *bytes.Buffer {
	var output bytes.Buffer

	for _, p := range job.ExcludeArguments() {
		output.WriteString(fmt.Sprintf("%s\n", p))
	}
	return &output
}

// InputPathListContents returns a *bytes.Buffer containing the contents of a
// input path list file. Does not write out the contents to a file. Returns
// (nil, nil) if there aren't any inputs without tickets associated with the
// Job.
func InputPathListContents(job *model.Job, pathListIdentifier, ticketsPathListIdentifier string) (*bytes.Buffer, error) {
	templateFields := OtherTemplateFields{
		PathListHeader:       pathListIdentifier,
		TicketPathListHeader: ticketsPathListIdentifier,
	}
	templateModel := TemplatesModel{
		job,
		templateFields,
	}

	if len(job.FilterInputsWithoutTickets()) > 0 {
		return generateFileContents(inputPathListTemplate, templateModel)
	}

	return nil, nil
}
