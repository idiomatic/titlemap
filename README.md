The titlemap tools are used to manage selection, naming, and (optionally)
parameters of videos for a transcoder such as HandBrakeCLI.

The title description file/stream has the format:

    Base Prefix | Title Number | Output Prefix | Transcoder Options

## example directly transcoding

    cat titles | titlemap ~/Movies --preset "$preset" $hbargs --outputdir ~/Public/ ~/Rip

## install

	go install ./cmd/...

## testing survey

	setopt +o nomatch
	preset="HQ 720p30 Surround"
	hbargs="--subtitle=none"
	collection="$preset $hbargs"
	ripdirs=( /Volumes/Archive*/Video/M*/*Rip
		/Volumes/Archive*/Video/TV*/*Rip{,/*Season*}
		/Volumes/Everything/Rip/*Rip{,/*Season*} )
	servedirs=( /Volumes/Galactica/Public/Video/M*/"$collection"
		 /Volumes/Galactica/Public/Video/TV*/"$collection"/* )
	reviewdir=~/Public/Review/"$collection"
	cat ~/titles | go run ~/go/src/github.com/idiomatic/titlemap/cmd/titlemap/main.go --color \
		 $ripdirs $servedirs $reviewdir

## transcoding

	# ... setup ...
    cat ~/titles | ~/go/bin/titlemap --progress --color \
		 --preset "$preset" $hbargs \
		 --outputdir $reviewdir \
	     $ripdirs $serverdirs $reviewdir
