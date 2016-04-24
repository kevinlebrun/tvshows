# www.pogdesign.co.uk/cat client

Fetch show that you followed or that you still didn’t watch.

## Usage

    $ go build -o shows *.go
    $ ./shows -password mypass -username myuser followed
    Continuum
    Falling Skies
    Hemlock Grove
    $ ./shows -password mypass -username myuser unwatched
    Continuum s04 e05 [The Desperate Hours]
    Continuum s04 e06 [Final Hour]
    Falling Skies s05 e10 [Reborn]
    Hemlock Grove s03 e09 [Damascus]
    Hemlock Grove s03 e10 [Brian's Song]

## License

The MIT license
