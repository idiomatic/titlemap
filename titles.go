package titlemap

import (
	"errors"
	"fmt"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
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

func clean(s string) string {
	s = strings.TrimSpace(s)
	if strings.ContainsRune(s, '\\') {
		s = `"` + s + `"`
		if unquoted, err := strconv.Unquote(s); err == nil {
			return unquoted
		}
	}
	// Be advised macOS prefers NFD (denormalized) filenames
	s, _, _ = transform.String(norm.NFD, s)
	return s
}

func (t *Title) Parse(line string) error {
	// escape # as \x23
	// escape | as \x7C
	t.RawComment = trimBefore(line, '#')
	line = trimAfter(line, '#')
	fields := splitOnRune(line, '|')

	fieldCount := len(fields)
	if fieldCount >= 1 {
		s := clean(fields[0])
		t.BaseName = s
		t.OutputBaseName = s
	}
	if fieldCount >= 2 {
		t.RawTranscodeArgs = strings.TrimSpace(fields[1])
	}
	if fieldCount >= 3 {
		t.OutputBaseName = clean(fields[2])
	}
	if fieldCount >= 4 {
		// XXX unused
		s := strings.TrimSpace(fields[3])
		if quoted, err := strconv.Unquote(s); err == nil {
			s = quoted
		}
		t.RawMetadata = s
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
