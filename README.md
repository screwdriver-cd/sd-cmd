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
$ sd-cmd exec namespace/name@version [arguments]
```

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
