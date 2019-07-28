package titlemap

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kballard/go-shellquote"
)

var (
	SourceExtensions = []string{".dvdmedia", ".mkv", ".bluray"}
	OutputExtensions = []string{".m4v", ".mp4"}

	ErrNotEnoughColumns   = errors.New("not enough columns")
	ErrInvalidTitleColumn = errors.New("blank or malformed title column")
	ErrNotExist           = os.ErrNotExist
)

func SplitOnRune(s string, sep rune) []string {
	return strings.FieldsFunc(s, func(c rune) bool {
		return c == sep
	})
}

func TrimAfter(s string, sep rune) string {
	if p := strings.IndexRune(s, sep); p > -1 {
		return s[0:p]
	}
	return s
}

func IsInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

type Title struct {
	SourceBaseName string // omitting SourceExtensions
	Title          string // blank or an int
	OutputBaseName string // omitting OutputExtensions
	TranscodeArgs  []string
}

func (t *Title) Parse(line string) error {
	line = TrimAfter(line, '#')
	fields := SplitOnRune(line, '|')
	if len(fields) < 3 {
		return ErrNotEnoughColumns
	}

	t.SourceBaseName = strings.TrimSpace(fields[0])
	t.Title = strings.TrimSpace(fields[1])
	t.OutputBaseName = strings.TrimSpace(fields[2])

	if len(fields) > 3 {
		args, err := shellquote.Split(strings.TrimSpace(fields[3]))
		if err != nil {
			return err
		}
		t.TranscodeArgs = args
	}

	// test late in case client feels it is a non-invalidating error
	if !IsInt(t.Title) {
		return ErrInvalidTitleColumn
	}

	return nil
}

func IsInvalid(err error) bool {
	return err == ErrNotEnoughColumns || err == ErrInvalidTitleColumn
}

// QueueSuffix encodes HandBrakeCLI arguments (if any).
func (t Title) QueueSuffix() string {

	var words []string

	if t.Title != "" {
		words = append(words, "--title", t.Title)
	}

	words = append(words, t.TranscodeArgs...)

	if len(words) > 0 {
		// trick a leading blank space
		words = append([]string{""}, words...)
	}

	return strings.Join(words, " ")
}

func (t Title) SourceExtension(dir string) (bool, string) {
	for _, ext := range SourceExtensions {
		_, err := os.Stat(filepath.Join(dir, t.SourceBaseName+ext))
		if err == nil {
			return true, ext
		}
	}
	return false, ""
}

func (t Title) SourceExistsIn(dir string) bool {
	found, _ := t.SourceExtension(dir)
	return found
}

func (t Title) OutputExtension(dir string) (bool, string) {
	for _, ext := range OutputExtensions {
		_, err := os.Stat(filepath.Join(dir, t.OutputBaseName+ext))
		if err == nil {
			return true, ext
		}
	}
	return false, ""
}

var destCache = make(map[string]map[string]string)

func ClearDestCache(root string) {
	destCache[root] = nil
}

// ScanDest walks once to build index of everything relevant.
func ScanDest(root string) error {
	if _, ok := destCache[root]; ok {
		return nil
	}

	d := make(map[string]string)
	destCache[root] = d

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		foundExt := filepath.Ext(path)
		for _, ext := range OutputExtensions {
			if foundExt == ext {
				base := strings.TrimSuffix(filepath.Base(path), ext)
				d[base] = filepath.Dir(path)
				break
			}
		}
		return nil
	})

	return err
}

func (t Title) Find(root, name string) (string, error) {
	if err := ScanDest(root); err != nil {
		return "", nil
	}

	dir, found := destCache[root][name]
	if !found {
		return "", ErrNotExist
	}

	return dir, nil
}

func (t Title) OutputExistsUnder(root string) (found bool) {
	_, err := t.Find(root, t.OutputBaseName)
	return err == nil
}
