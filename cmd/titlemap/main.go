package main

// Transcode missing videos using titlemap input
//
// Stdin specifies the "titlemap" transcode parameters.  This includes
// input and output filename bases (i.e., omitting directory and
// extension), the title number, and other HandBrake parameters.
//
// The command line arguments list directories with videos.
// Input videos are skipped if they are not found.
// Output videos are skipped if they are found.

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/idiomatic/titlemap"
)

const (
	outputExt                   = ".m4v"
	logExt                      = ".log"
	outputDirAdjacentLogArchive = ",log"
	listenAddr                  = ":8888"
)

// constant-ish, i.e., no sync required
var (
	showProgress     bool
	quiet            bool
	outputDir        string
	logDir           string
	txArgs           TranscodeArgs // common transcode args
	inputExtensions  = []string{".dvdmedia", ".mkv", ".bluray"}
	outputExtensions = []string{".m4v", ".mp4"}
)

// dynamic
var (
	doneMutex        sync.Mutex
	done             bool
	ctx, cancel      = context.WithCancel(context.Background())
	currentMutex     sync.Mutex
	currentTranscode *Transcode
	currentProgress  *TranscodeProgress
	currentStarted   time.Time
	currentTitle     *titlemap.Title
)

func since(start time.Time) Duration {
	return Duration(time.Since(start).Round(time.Second))
}

func setDone() {
	doneMutex.Lock()
	defer doneMutex.Unlock()
	done = true
}

func isDone() bool {
	doneMutex.Lock()
	defer doneMutex.Unlock()
	return done
}

func main() {
	// parse command line switches
	flag.StringVar(&outputDir, "outputdir", outputDir, "Dir for transcode output")
	flag.StringVar(&logDir, "logdir", logDir, "Dir for transcode logs")
	flag.BoolVar(&showProgress, "progress", showProgress, "Transcode progress")
	flag.BoolVar(&quiet, "quiet", color, "Hide inactionable output")
	flag.BoolVar(&color, "color", color, "Output with colors")
	txArgs.DefineFlags(flag.CommandLine)
	flag.Parse()

	inputs := make(map[string]string)
	outputs := make(map[string]struct{})

	// scan dirs for different video categories
	// HACK dir is both an source and destination
	for _, dir := range flag.Args() {
		// HACK skip non-directories
		if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
			continue
		}

		names, err := Readdirnames(dir)
		if err != nil {
			log.Fatal(err)
		}

		inputNames := FilterSuffixes(names, inputExtensions)
		for _, name := range inputNames {
			base := strings.TrimSuffix(name, filepath.Ext(name))
			if _, found := inputs[base]; !found {
				// first wins
				inputs[base] = filepath.Join(dir, name)
			}
		}

		outputNames := FilterSuffixes(names, outputExtensions)
		for _, name := range outputNames {
			base := strings.TrimSuffix(name, filepath.Ext(name))
			outputs[base] = struct{}{}
		}
	}

	// pick where to archive logs
	if logDir == "" {
		logDir = filepath.Join(outputDir, outputDirAdjacentLogArchive)
		if info, err := os.Stat(logDir); err != nil || !info.IsDir() {
			logDir = outputDir
		}
	}

	// cleanup and exit on Ctrl-C
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		log.Printf("quitting after this transcode\n")
		setDone()
		<-signalChan
		log.Printf("quitting now\n")
		cancel()
	}()

	// process lines of titlemap-formatted standard input
	scanner := bufio.NewScanner(os.Stdin)

	for !isDone() {
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()

		t := titlemap.Title{}
		if err := t.Parse(line); titlemap.IsInvalid(err) {
			continue
		} else if err != nil {
			log.Fatal(err)
		}

		// skip if already processed
		if _, found := outputs[t.OutputBaseName]; found {
			if !quiet {
				log.Println(ProcessedTitle(t))
			}
			continue
		}

		// skip if input is offline
		input, found := inputs[t.BaseName]
		if !found {
			if !quiet {
				log.Println(OfflineTitle(t))
			}
			continue
		}

		// copy and modify
		args := txArgs
		err := args.Parse(t.TranscodeArgs())
		if err != nil {
			log.Fatal(err)
		}

		// transcode trigger
		if outputDir != "" {
			finalOutput := filepath.Join(outputDir, t.OutputBaseName+outputExt)
			finalLog := filepath.Join(logDir, t.OutputBaseName+outputExt+logExt)

			log.Println(ProcessingTitle(t, args.Args()))

			tCopy := t
			currentMutex.Lock()
			currentTitle = &tCopy
			currentStarted = time.Now()
			currentMutex.Unlock()

			err := transcodeHelper(input, args, finalOutput, finalLog)
			if err != nil {
				log.Fatal(err)
			}

			currentMutex.Lock()
			elapsed := since(currentStarted)
			currentStarted = time.Time{}
			currentMutex.Unlock()

			log.Println(JustProcessedTitle(t, time.Duration(elapsed)))
		} else {
			if !quiet {
				log.Println(TodoTitle(t))
			}
		}
	}
}

// transcodeHelper sets up, atomically finalizes, and cleans up after
// a transcode or transcode attempt.  It is a discrete function to
// trigger defer after each titlemap line.
func transcodeHelper(input string, txArgs TranscodeArgs, finalOutput string, finalLog string) (err error) {
	// pick a temporary name
	tempOutput, err := TempFile(finalOutput)
	if err != nil {
		return err
	}

	// clean up in case of failure
	defer os.Remove(tempOutput)

	tmpLog := tempOutput + logExt
	logFile, err := os.OpenFile(tmpLog, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	// XXX possibly redundant Close()
	defer logFile.Close()

	// clean up in case of failure before Rename
	defer os.Remove(tmpLog)

	tx := Transcode{
		Input:   input,
		Args:    txArgs,
		Output:  tempOutput,
		Log:     logFile,
		Context: ctx,
	}

	if showProgress {
		// client creates channel so it can decide on buffering
		tx.ProgressChan = make(chan TranscodeProgress, 1)

		go func() {
			for tp := range tx.ProgressChan {
				//fmt.Printf("\r%.1f%% %s\033[K", ts.Progress*100, ts.ETA)
				fmt.Print("\r", tp.Raw, clearToEndOfLine())

				tpCopy := tp
				currentMutex.Lock()
				currentProgress = &tpCopy
				currentMutex.Unlock()
			}
		}()
	}

	tCopy := tx
	currentMutex.Lock()
	currentTranscode = &tCopy
	currentMutex.Unlock()

	defer func() {
		currentMutex.Lock()
		currentTranscode = nil
		currentProgress = nil
		currentMutex.Unlock()
	}()

	if err = tx.Run(); err != nil {
		return err
	}

	if err = os.Rename(tempOutput, finalOutput); err != nil {
		return err
	}

	if err = os.Rename(tmpLog, finalLog); err != nil {
		return err
	}

	return nil
}

// TempFile generates a unique filename near fullPath.
func TempFile(fullPath string) (string, error) {
	dir, file := filepath.Split(fullPath)
	ext := filepath.Ext(file)
	base := strings.TrimSuffix(file, ext)
	placeholder, err := ioutil.TempFile(dir, base+"-*"+ext)
	if err != nil {
		return "", err
	}
	placeholder.Close()

	return placeholder.Name(), nil
}
