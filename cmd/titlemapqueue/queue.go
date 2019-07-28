package main

// Symlink into transcode queue directory source video from source dir(s).
//
// Standard input is the list of video "tracks" in the format:
//     source_name | title_number | output_name | extra_transcode_args

import (
	"bufio"
	"flag"
	"os"
	"path/filepath"

	"github.com/idiomatic/titlemap"
)

var (
	queueDir   string
	sourceDirs = []string{"./"}
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	queueDir = filepath.Join(home, "Queue")
}

func main() {
	flag.StringVar(&queueDir, "queue", queueDir, "Dir for queueing DVD track transcodes")
	flag.Parse()
	sourceDirs = flag.Args()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		var t titlemap.Title
		if err := t.Parse(line); titlemap.IsInvalid(err) {
			continue
		} else if err != nil {
			panic(err)
		}

		suffix := t.QueueSuffix()

		for _, sourceDir := range sourceDirs {
			found, ext := t.SourceExtension(sourceDir)
			if found {
				oldName, err := filepath.Abs(filepath.Join(sourceDir, t.SourceBaseName+ext))
				if err != nil {
					panic(err)
				}

				// XXX  if t.OutputBaseName == "" ...

				newName := filepath.Join(queueDir, t.OutputBaseName+suffix+ext)

				err = os.Symlink(oldName, newName)
				if err != nil && !os.IsExist(err) {
					panic(err)
				}

				break
			}
		}
	}
}
