package main

// Filter input, omitting videos that exist somewhere under destroot(s).
//
// Standard input is the list of video "tracks" in the format:
//     source_name | title_number | output_name | extra_transcode_args

import (
	"bufio"
	"os"

	"github.com/idiomatic/titlemap"
)

var (
	destRoots = []string{"."}
)

func main() {
	out := bufio.NewWriter(os.Stdout)
	linesSinceFlush := 0
	defer out.Flush()

	destRoots = os.Args[1:]

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()

		var t titlemap.Title
		if err := t.Parse(line); titlemap.IsInvalid(err) {
			continue
		} else if err != nil {
			panic(err)
		}

		var found bool
		if t.OutputBaseName != "" {
			for _, destRoot := range destRoots {
				if _, err := t.Find(destRoot, t.OutputBaseName); err == nil {
					found = true
					break
				}
			}
		}

		if !found {
			out.Write([]byte(line))
			out.WriteRune('\n')
			linesSinceFlush++

			// courtesy flush to approximately align flushes with newlines
			if linesSinceFlush >= 20 {
				out.Flush()
				linesSinceFlush = 0
			}
		}
	}
}
