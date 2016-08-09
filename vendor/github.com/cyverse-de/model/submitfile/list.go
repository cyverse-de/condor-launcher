package submitfile

import (
	"bytes"
	"regexp"
)

var escapeCharsRegexp = regexp.MustCompile(`[\\\s"']`)

// escapeChar escapes a single character in a string literal with an encoded version of that
// character to be placed in a string literal in an HTCondor submit file. Whitespace characters
// are replaced by their corresponding metacharacters. Quotes, apostrophes and backslashes are
// escaped using backslashes.
func escapeChar(orig string) string {
	switch orig {
	case "\t":
		return `\t`
	case "\n":
		return `\n`
	case "\f":
		return `\f`
	case "\r":
		return `\r`
	case `"`:
		return `\"`
	case `'`:
		return `\'`
	case `\`:
		return `\\`
	default:
		return orig
	}
}

// FormatList converts a slice of strings to a formatted list that can be placed in
// an HTCondor submit file.
func FormatList(l []string) string {
	result := bytes.Buffer{}

	result.WriteRune('{')
	for index, group := range l {
		if index > 0 {
			result.WriteRune(',')
		}
		result.WriteRune('"')
		result.WriteString(escapeCharsRegexp.ReplaceAllStringFunc(group, escapeChar))
		result.WriteRune('"')
	}
	result.WriteRune('}')

	return result.String()
}
