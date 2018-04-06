package model

import (
	"fmt"
	"path"
	"strings"
)

// StepInput describes a single input for a job step.
type StepInput struct {
	ID           string `json:"id"`
	Ticket       string `json:"ticket"`
	Multiplicity string `json:"multiplicity"`
	Name         string `json:"name"`
	Property     string `json:"property"`
	Retain       bool   `json:"retain"`
	Type         string `json:"type"`
	Value        string `json:"value"`
}

// IRODSPath returns a string containing the iRODS path to an input file.
func (i *StepInput) IRODSPath() string {
	if i.Multiplicity == "collection" {
		if !strings.HasSuffix(i.Value, "/") {
			return fmt.Sprintf("%s/", i.Value)
		}
	}
	return i.Value
}

// Identifier returns a string containing the input job's identifier in the
// format "input-<suffix>"
func (i *StepInput) Identifier(suffix string) string {
	return fmt.Sprintf("input-%s", suffix)
}

// Stdout returns a string containing the path to the input job's stdout file.
// It should be a relative path in the format "logs/logs-stdout-<i.Identifier(suffix)>"
func (i *StepInput) Stdout(suffix string) string {
	return path.Join("logs", fmt.Sprintf("logs-stdout-%s", i.Identifier(suffix)))
}

// Stderr returns a string containing the path to the input job's stderr file.
// It should be a relative path in the format "logs/logs-stderr-<i.Identifier(suffix)>"
func (i *StepInput) Stderr(suffix string) string {
	return path.Join("logs", fmt.Sprintf("logs-stderr-%s", i.Identifier(suffix)))
}

// LogPath returns the path to the Condor log file for the input job. The returned
// path will be in the format "<parent>/logs/logs-condor-<i.Identifier(suffix)>"
func (i *StepInput) LogPath(parent, suffix string) string {
	return path.Join(parent, "logs", fmt.Sprintf("logs-condor-%s", i.Identifier(suffix)))
}

// Source returns the path to the local filename of the input file.
func (i *StepInput) Source() string {
	value := path.Base(i.Value)
	if i.Multiplicity == "collection" {
		if !strings.HasSuffix(value, "/") {
			return fmt.Sprintf("%s/", value)
		}
	}
	return value
}

// Arguments returns the porklock settings needed for the input operation.
func (i *StepInput) Arguments(username string, metadata []FileMetadata) []string {
	path := i.IRODSPath()
	args := []string{
		"get",
		"--user", username,
		"--source", path,
	}
	for _, m := range MetadataArgs(metadata).FileMetadataArguments() {
		args = append(args, m)
	}
	return args
}

// StepOutput describes a single output for a job step.
type StepOutput struct {
	Multiplicity string `json:"multiplicity"`
	Name         string `json:"name"`
	Property     string `json:"property"`
	QualID       string `json:"qual-id"`
	Retain       bool   `json:"retain"`
	Type         string `json:"type"`
}

// Source returns the path to the local filename for the output file.
func (o *StepOutput) Source() string {
	value := o.Name
	if o.Multiplicity == "collection" {
		if !path.IsAbs(value) {
			value = fmt.Sprintf("/de-app-work/%s", value)
		}
		if !strings.HasSuffix(value, "/") {
			value = fmt.Sprintf("%s/", value)
		}
	}
	return value
}
