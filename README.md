# sd-cmd
[![Build Status][build-image]][build-url]
[![Go Report Card][goreport-image]][goreport-url]  
Screwdriver commands (Sharing Binaries)

## Usage

### Install
```bash
$ go get github.com/screwdriver-cd/sd-cmd
$ cd $GOPATH/src/github.com/screwdriver-cd/sd-cmd
$ go build -a -o sd-cmd
```

### Execute
```bash
$ sd-cmd exec [flags] namespace/name@version [arguments]

# usage
# Flags:
#   -debug, --debug    output debug logs to a file
```

#### Debug mode
In debug mode, the debug log can be output to a file.  
It can be used in one of the following ways.
- Use `-debug` of `--debug` option
- Set `SD_CMD_DEBUG_LOG` environment variable to `true`

## Testing
```bash
go get github.com/screwdriver-cd/sd-cmd
go test -cover github.com/screwdriver-cd/sd-cmd/...
```

## License
Code licensed under the BSD 3-Clause license. See LICENSE file for terms.

[build-image]: https://cd.screwdriver.cd/pipelines/408/badge
[build-url]: https://cd.screwdriver.cd/pipelines/408
[goreport-image]: https://goreportcard.com/badge/github.com/Screwdriver-cd/sd-cmd
[goreport-url]: https://goreportcard.com/report/github.com/Screwdriver-cd/sd-cmd
