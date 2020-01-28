package main

import (
	"testing"
)

func TestParse(t *testing.T) {
	ts := TranscodeStatus{}
	err := ts.Parse("Encoding: task 1 of 2, 2.50 % (0.00 fps, avg 0.00 fps, ETA 00h00m38s)")
	if err != nil {
		t.Error(err)
	}
	if ts.Task != 1 {
		t.Error("task")
	}
}
