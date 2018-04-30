package jobs

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"os"
	"path"
	"text/template"
)

func generateFile(dirPath, filename string, t *template.Template, data interface{}) (string, error) {

	// Open the output file for writing.
	filePath := path.Join(dirPath, filename)
	out, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", errors.Wrapf(err, "unable to open %s for output", filePath)
	}
	defer out.Close()

	// Generate the file from the template.
	err = t.Execute(out, data)
	if err != nil {
		return "", errors.Wrapf(err, "unable to generate %s from template %s", filePath, t.Name())
	}

	return filePath, nil
}

func generateFileContents(t *template.Template, data interface{}) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	err := t.Execute(&buffer, data)
	if err != nil {
		return &buffer, errors.Wrapf(err, "failed to apply data to the %s template", t.Name())
	}
	return &buffer, err
}

func generateJson(dirPath, filename string, data interface{}) (string, error) {

	// Open the output file for writing.
	filePath := path.Join(dirPath, filename)
	out, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", errors.Wrapf(err, "unable to open %s for output", filePath)
	}
	defer out.Close()

	// Generate the output JSON.
	encoder := json.NewEncoder(out)
	err = encoder.Encode(data)
	if err != nil {
		return "", errors.Wrapf(err, "unable to marshal JSON to %s", filePath)
	}

	return filePath, nil
}

func generateJsonContents(data interface{}) (*bytes.Buffer, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(data)
	if err != nil {
		return &buffer, errors.Wrap(err, "unable to marshal JSON")

	}

	return &buffer, nil
}
