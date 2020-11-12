The titlemap tools are used to manage a video library.

Titlemap succinctly declares "input" video sources and their
conversion into "outputs".

Outputs should be unique (within the outputdir).  However, inputs may
be non-unique, as would be the case of a DVD containing multiple
episodes.  In the case of non-unique input, the transcode-args column
would likely make it result in differing output.

The titlemap lines are succinct for a reason.  Directory prefix and
filename extensions are omitted to permit "multiple passes" of the
same mapping.  The outputdir directory prefix are used to isolate
different global transcoding parameters.  Furthermore, different
transcoding parameters may also result in different file formats,
i.e. extensions.  Also, the specification of just a title is such a
common case that the transcode-args column is supports the shorthand.

---

The title description file/stream has the format:

    Base Prefix | Title Number Or Transcode Options | Output Prefix | Metadata

## example transcoding

    titlemap --preset="$preset" --outputdir=~/Movies/New ~/Movies ~/Rip < titles

## install

    go install ./cmd/...

## survey example

    setopt nullglob
    preset="HQ 720p30 Surround"
    hbargs="--subtitle=none"
    collection="$preset $hbargs"
    servedirs=(
        /Volumes/Galactica/Public/Video/M*/"$collection"
        /Volumes/Galactica/Public/Video/TV*/"$collection"/*
    )
    reviewdir=~/Public/Review/"$collection"
    cat ~/titles | ~/go/bin/titlemap --color \
        /Volumes/Archive*/Video/M*/*Rip \
        /Volumes/Archive*/Video/TV*/*Rip{,/*Season*} \
        /Volumes/Rocinate/Rip/*Rip{,/*Season*} \
        $servedirs $reviewdir

## transcoding examples

### Very Fast 720p "H.264"

    (
        setopt nullglob;
        preset="Very Fast 720p30"
        hbargs=(--subtitle=none)
        collection="$preset $hbargs"
        servedirs=(/Volumes/Galactica/Public/Video/{Mov*/"$collection",TV*/"$collection"/*})
        reviewdir=~/Public/Review/"$collection"
        ripdirs=(/Volumes/{{PlanetExpress,Rocinate}/Archive,Archive*}/Video/{Mov*/*Rip,TV*/*Rip{,/*Season*}} /Volumes/Rocinate/Rip/*Rip{,/*Season*})
        cat ~/titles | ~/go/bin/titlemap --progress --color \
            --preset "$preset" $hbargs \
            --outputdir "$reviewdir" \
            $ripdirs $servedirs $reviewdir
    )

### Very Fast 480p "H.265"

    ...
        preset="Very Fast 480p30"
        hbargs=(--encoder=x265_10bit --subtitle=none)
    ...

### HQ 720p "H.264"

    ...
        preset="HQ 720p30 Surround"
        hbargs=(--subtitle=none)
    ...

### HQ 1080p "H.265"

    ...
        preset="HQ 1080p30 Surround"
        hbargs=(--encoder=x265_10bit --subtitle=none)
    ...

### Manually

    (
        title="..."
        title="Once Upon a Time in Hollywood"
        setopt nullglob
        ripdir=/Volumes/Rocinate/Rip/BluRay\ Rip
        for args in \
            'Very Fast 720p30 --subtitle=none' \
            'Very Fast 480p30 --encoder=x265_10bit --subtitle=none' \
            'HQ 720p30 Surround --subtitle=none' \
            'HQ 1080p30 Surround --encoder=x265_10bit --subtitle=none'
        do
			preset=${args%% --*} hbargs=${args#$preset }
			out=~/Public/Review/"$args/$title.m4v"
            [ -e "$out" ] || HandBrakeCLI --preset "$preset" $hbargs --input "$ripdir/$title.mkv" --output "$out"
        done
    )

## monitoring webservice

    curl localhost:8888

