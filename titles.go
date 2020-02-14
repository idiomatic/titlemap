package titlemap

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kballard/go-shellquote"
)

var ErrInvalidTitleColumn = errors.New("blank or malformed title column")

func splitOnRune(s string, sep rune) []string {
	return strings.FieldsFunc(s, func(c rune) bool {
		return c == sep
	})
}

func trimAfter(s string, sep rune) string {
	if p := strings.IndexRune(s, sep); p > -1 {
		return s[0:p]
	}
	return s
}

func trimBefore(s string, sep rune) string {
	if p := strings.IndexRune(s, sep); p > -1 {
		return s[p:]
	}
	return ""
}

func isInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

type Title struct {
	BaseName         string `json:"base_name"` // omitting extensions
	RawTranscodeArgs string `json:"raw_transcode_args,omitempty"`
	OutputBaseName   string `json:"output_base_name,omitempty"` // alternate unique base, omitting extensions
	RawMetadata      string `json:"raw_metadata,omitempty"`
	RawComment       string `json:"raw_comment,omitempty"`
}

func (t Title) String() string {
	return fmt.Sprintf("%s|%s|%s|%s%s",
		t.BaseName, t.RawTranscodeArgs, t.OutputBaseName, t.RawMetadata, t.RawComment)
}

func (t *Title) Parse(line string) error {
	t.RawComment = trimBefore(line, '#')
	line = trimAfter(line, '#')
	fields := splitOnRune(line, '|')

	fieldCount := len(fields)
	if fieldCount >= 1 {
		t.BaseName = strings.TrimSpace(fields[0])
		t.OutputBaseName = t.BaseName
	}
	if fieldCount >= 2 {
		t.RawTranscodeArgs = strings.TrimSpace(fields[1])
	}
	if fieldCount >= 3 {
		t.OutputBaseName = strings.TrimSpace(fields[2])
	}
	if fieldCount >= 4 {
		t.RawMetadata = strings.TrimSpace(fields[3])
	}

	return nil
}

func (t Title) TranscodeArgs() []string {
	args, _ := shellquote.Split(strings.TrimSpace(t.RawTranscodeArgs))

	// shorthand for title
	if isInt(t.RawTranscodeArgs) {
		return []string{"--title=" + t.RawTranscodeArgs}
	}

	return args
}

func IsInvalid(err error) bool {
	return err == ErrInvalidTitleColumn
}
