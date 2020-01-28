package main

import (
	"os"
	"strings"
)

func Readdirnames(dir string) ([]string, error) {
	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return f.Readdirnames(-1)
}

func FilterSuffixes(names []string, suffixes []string) []string {
	var filtered []string
	for _, name := range names {
		for _, ext := range suffixes {
			if strings.HasSuffix(name, ext) {
				filtered = append(filtered, name)
				break
			}
		}
	}
	return filtered
}
