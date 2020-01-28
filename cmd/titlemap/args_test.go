package main

import (
	"flag"
	"testing"
)

func TestFlagParse(t *testing.T) {
	var title string
	var preset string

	f := flag.NewFlagSet("HandBrakeCLI", flag.ContinueOnError)
	f.StringVar(&title, "title", title, "HandBrake title")
	f.StringVar(&preset, "preset", preset, "HandBrake preset")
	err := f.Parse([]string{"--title=42", "-preset", "HQ 720p30 Surround"})
	if err != nil {
		t.Error(err)
	}
	if title != "42" {
		t.Error("title not 42")
	}
	if preset != "HQ 720p30 Surround" {
		t.Error("preset not valid")
	}
}

func TestTranscodeArgsParse(t *testing.T) {
	var txArgs TranscodeArgs

	err := txArgs.Parse([]string{"--title=42"})
	if err != nil {
		t.Error(err)
	}
	if txArgs.Title != 42 {
		t.Error(txArgs.Title)
	}
}
