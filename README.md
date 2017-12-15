# sd-cmd
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
$ sd_cmd exec namespace/command@version [arguments]
```

## Testing
Some tests create directory in your computer. Therefore there is possibility to fail test by permission error.
```bash
go get github.com/screwdriver-cd/sd-cmd
go test -cover github.com/screwdriver-cd/sd-cmd/...
```

## License
Code licensed under the BSD 3-Clause license. See LICENSE file for terms.