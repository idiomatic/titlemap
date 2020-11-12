package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Transcode struct {
	Input        string                 `json:"input"`
	Output       string                 `json:"output"`
	Args         TranscodeArgs          `json:"args"`
	ProgressChan chan TranscodeProgress `json:"-"`
	Log          io.Writer              `json:"-"`
	Context      context.Context        `json:"-"`
}

type TranscodeProgress struct {
	Task       int      `json:"task"`
	Tasks      int      `json:"tasks"`
	Progress   Progress `json:"progress"`
	FPS        float32  `json:"fps"`
	AverageFPS float32  `json:"average_fps"`
	ETA        Duration `json:"eta"`
	Raw        string   `json:"raw"`
}

// from 0.0 to 1.0
type Progress float32

func (p Progress) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Normalized float32 `json:"normalized"`
		Percent    string  `json:"percent"`
	}{float32(p), fmt.Sprintf("%.1f%%", float32(p)*100)})
}

type Duration time.Duration

func (ts Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Nanoseconds time.Duration `json:"nanoseconds,omitempty"`
		Human       string        `json:"human,omitempty"`
	}{time.Duration(ts), time.Duration(ts).String()})
}

func (ts Duration) String() string {
	return time.Duration(ts).String()
}

// TranscodeArgs specifies popular HandBrake options
type TranscodeArgs struct {
	Preset       string `json:"preset,omitempty"`
	Title        int    `json:"title,omitempty"`
	Chapters     string `json:"chapters,omitempty"`
	StartAt      string `json:"start_at,omitempty"`
	StopAt       string `json:"stop_at,omitempty"`
	Encoder      string `json:"encoder,omitempty"`
	Subtitle     string `json:"subtitle,omitempty"`
	NLMeans      string `json:"nlmeans,omitempty"`
	Crop         string `json:"crop,omitempty"`
	Audio        string `json:"audio,omitempty"`
	AudioEncoder string `json:"audio_encoder,omitempty"`
}

// Args returns a slice of non-default HandBrakeCLI args
func (ta TranscodeArgs) Args() []string {
	var args []string
	addStringArg := func(arg, value string) {
		if value != "" {
			args = append(args, arg, value)
		}
	}
	addIntArg := func(arg string, value int) {
		if value != 0 {
			args = append(args, arg, strconv.Itoa(value))
		}
	}
	addStringArg("--preset", ta.Preset)
	addIntArg("--title", ta.Title)
	addStringArg("--chapters", ta.Chapters)
	addStringArg("--start-at", ta.StartAt)
	addStringArg("--stop-at", ta.StopAt)
	addStringArg("--encoder", ta.Encoder)
	addStringArg("--subtitle", ta.Subtitle)
	addStringArg("--nlmeans", ta.NLMeans)
	addStringArg("--crop", ta.Crop)
	addStringArg("--audio", ta.Audio)
	addStringArg("--aencoder", ta.AudioEncoder)
	return args
}

// DefineFlags registers popular HandBrake parameters as flags.  Used
// on command line args and the titlemap transcodeArgs column.
func (ta *TranscodeArgs) DefineFlags(f *flag.FlagSet) {
	f.StringVar(&ta.Preset, "preset", ta.Preset,
		"HandBrake preset")
	f.IntVar(&ta.Title, "title", ta.Title, "HandBrake title")
	f.StringVar(&ta.Chapters, "chapters", ta.Chapters,
		"HandBrake chapters")
	f.StringVar(&ta.StartAt, "start-at", ta.StartAt,
		"HandBrake start-at offset")
	f.StringVar(&ta.StopAt, "stop-at", ta.StopAt,
		"HandBrake stop-at duration after start-at")
	f.StringVar(&ta.Encoder, "encoder", ta.Encoder,
		"HandBrake video encoder")
	f.StringVar(&ta.Subtitle, "subtitle", ta.Subtitle,
		"HandBrake subtitle track")
	f.StringVar(&ta.NLMeans, "nlmeans", ta.NLMeans,
		"HandBrake video denoise filter")
	f.StringVar(&ta.Crop, "crop", ta.Crop,
		"HandBrake picture cropping")
	f.StringVar(&ta.Audio, "audio", ta.Audio,
		"HandBrake audio")
	f.StringVar(&ta.AudioEncoder, "aencoder", ta.AudioEncoder,
		"HandBrake audio encoder")
}

// Parse a titlemap transcodeArgs column.
func (ta *TranscodeArgs) Parse(args []string) error {
	f := flag.NewFlagSet("HandBrakeCLI", flag.ContinueOnError)
	ta.DefineFlags(f)
	err := f.Parse(args)
	if err != nil {
		return err
	}
	if len(f.Args()) > 0 {
		return errors.New("transcode flags error")
	}
	return nil
}

// Run HandBrake to transcode a video.
func (t *Transcode) Run() error {
	args := []string{"--input", t.Input, "--output", t.Output}
	args = append(args, t.Args.Args()...)

	ctx := t.Context
	if ctx == nil {
		ctx = context.Background()
	}

	cmd := exec.CommandContext(ctx, "HandBrakeCLI", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	cmd.Stderr = t.Log

	progress, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if t.ProgressChan != nil {
		defer close(t.ProgressChan)
	}
	go func() {
		for {
			b := make([]byte, 80)
			n, err := progress.Read(b)
			if err == io.EOF {
				break
			} else if err != nil {
				// XXX log.Println("reader error", err)
			}
			ts := &TranscodeProgress{}
			err = ts.Parse(strings.TrimPrefix(string(b[:n]), "\r"))
			if t.ProgressChan != nil {
				// non-blocking
				select {
				case t.ProgressChan <- *ts:
				default:
				}
			}
		}
	}()
	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

// Parse HandBrake progress output.
func (ts *TranscodeProgress) Parse(line string) error {
	var eta string

	ts.Raw = line
	_, err := fmt.Sscanf(line, "Encoding: task %d of %d, %f %% (%f fps, avg %f fps, ETA %s", &ts.Task, &ts.Tasks, &ts.Progress, &ts.FPS, &ts.AverageFPS, &eta)
	if err != nil {
		return err
	}
	ts.Progress /= 100.0
	// HACK scanning "%s)" does not exclude paren from string
	eta = strings.TrimSuffix(eta, ")")
	parsedETA, err := time.ParseDuration(eta)
	ts.ETA = Duration(parsedETA)

	return nil
}
