package main

// XXX import "github.com/gdamore/tcell"

import (
	"fmt"
	"time"

	"github.com/idiomatic/titlemap"
	"github.com/kballard/go-shellquote"
)

var color bool

func sgrSpan(attr, s string) string {
	if color {
		return fmt.Sprintf("\033[%sm%s\033[m", attr, s)
	}
	return s
}

func clearToEndOfLine() string {
	if color {
		return "\033[K"
	}
	return ""
}

func red(s string) string {
	return sgrSpan("31", s)
}

func green(s string) string {
	return sgrSpan("32", s)
}

func yellow(s string) string {
	return sgrSpan("33", s)
}

func magenta(s string) string {
	return sgrSpan("35", s)
}

func gray(s string) string {
	return sgrSpan("37", s)
}

func ProcessedTitle(t titlemap.Title) string {
	return fmt.Sprintf("%s|%s|%s|%s%s",
		t.BaseName, t.RawTranscodeArgs,
		green(t.OutputBaseName), t.RawMetadata, gray(t.RawComment))
}

func OfflineTitle(t titlemap.Title) string {
	return fmt.Sprintf("%s|%s|%s|%s%s",
		red(t.BaseName), t.RawTranscodeArgs,
		t.OutputBaseName, t.RawMetadata, gray(t.RawComment))
}

func TodoTitle(t titlemap.Title) string {
	return fmt.Sprintf("%s|%s|%s|%s%s",
		t.BaseName, t.RawTranscodeArgs,
		yellow(t.OutputBaseName), t.RawMetadata, gray(t.RawComment))
}

func ProcessingTitle(t titlemap.Title, args []string) string {
	return fmt.Sprintf("%s|%s|%s|%s%s",
		yellow(t.BaseName), yellow(shellquote.Join(args...)),
		yellow(t.OutputBaseName), t.RawMetadata, gray(t.RawComment))
}

func JustProcessedTitle(t titlemap.Title, elapsed time.Duration) string {
	return fmt.Sprintf("%s # %s", ProcessedTitle(t), magenta(elapsed.String()))
}
