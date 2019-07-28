The titlemap tools are used to manage selection, naming, and (optionally)
parameters of source videos for a transcoder such as HandBrakeCLI.

Per the UNIX philosophy, they do one thing and do it well.  They are
designed for composition by passing state via a standard I/O pipeline.

The title description file/stream is:

    Source Prefix | Title Number | Output Prefix [ | Transcoder Options ]

## queueing for transcode

    cat titles | titlemapomit ~/Movies | titlemapqueue --queue ~/Queue/Very\ Fast\ 720p30 ~/Rip

## fixing names after LaunchAgent

Removes `--title` and other transcoder filename-encoded command-line switches.

    for n in *--*; do mv -i "$n" "${n%% --*}.${n##*.}"; done
