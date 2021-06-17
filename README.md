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
We execute a command. The command arguments can be specified by the following `arguments`.
```bash
$ sd-cmd exec [flags] namespace/name@version [arguments]

# usage
# Flags:
#   -debug, --debug    Output debug logs to a file
```

#### Debug mode
In debug mode, the debug log can be output to a file.  
It can be used in one of the following ways.
- Use `-debug` of `--debug` option
- Set `SD_CMD_DEBUG_LOG` environment variable to `true`

### Publish
We publish the command specified by a yaml.
```bash
$ sd-cmd publish [flags]

# usage
# Flags:
#   -f, --f    Specify the path of yaml to publish (default: sd-command.yaml)
#   -t, --t    Specify the tag given to the command (default: sd-command.yaml)
```

### Validate
We validate that a yaml is in the correct format for the command.
```bash
$ sd-cmd validate [flags]

# usage
# Flags:
#   -f, --f    Specify the path of yaml to validate (default: sd-command.yaml)
```

### Promote
We give the tag specified by `tag` to the version specified by `targetVersion`. The tag will be removed from the version which `tag` is originally assigned to.
```bash
$ sd-cmd promote namespace/name targetVersion tag
```

### Remove tag
We remove the tag specified by `tag` from the version which `tag` is assigned to.
```bash
$ sd-cmd removeTag namespace/name tag
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
