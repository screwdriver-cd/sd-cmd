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
Executing a published command. The arguments for published command can be specified by the following `arguments`.
```bash
USAGE:
   sd-cmd exec [options] [namespace/name@version] [arguments...]

OPTIONS:
   -debug, --debug    Output debug logs to a file
   -v, --v            Output verbose log to console

EXAMPLE:
   sd-cmd exec foo/bar@stable arg1 arg2
```

#### Debug mode
In debug mode, the debug log can be output to a file.  
It can be used in one of the following ways.
- Use `-debug` of `--debug` option
- Set `SD_CMD_DEBUG_LOG` environment variable to `true`

### Validate
Validating if yaml is correct as sd-cmd format.
```bash
USAGE:
   sd-cmd validate [options]

OPTIONS:
   -f, --f string    Specify the path of yaml to validate (default: sd-command.yaml)

EXAMPLE:
   sd-cmd validate -f ./sd-command.yaml
```

### Publish
Publishing a command specified by a yaml.
```bash
USAGE:
   sd-cmd publish [options]

OPTIONS:
   -f, --f string    Specify the path of yaml to publish (default: sd-command.yaml)
   -t, --t string    Specify the tag given to the command (default: latest)

EXAMPLE:
   sd-cmd publish -f ./sd-command.yaml -t latest
```

### Promote
Giving a `tag` to a `targetVersion` of command. If a `tag` is already set to another version, that tag will be moved to `targetVersion`. `targetVersion` can be set exact version or tag (e.g. 1.0.1, latest).
```bash
USAGE:
   sd-cmd promote [namespace/name] [targetVersion] [tag]

EXAMPLE:
   sd-cmd promote foo/bar latest stable
   sd-cmd promote foo/bar 1.0.1 stable
```

### Remove tag
Removing a `tag` from a version of published command.
```bash
USAGE:
   sd-cmd removeTag [namespace/name] [tag]

EXAMPLE:
   sd-cmd removeTag foo/bar stable
```

## Testing
```bash
go get github.com/screwdriver-cd/sd-cmd
go test -cover github.com/screwdriver-cd/sd-cmd/...
```

## Local configuration
We can run sd-cmd locally by setting the environment variable as follows.
```bash
$ export SD_API_URL=${YOUR_SD_API_HOST}/v4/
$ export SD_STORE_URL=${YOUR_SD_STORE_HOST}/v1/
$ export SD_TOKEN=${YOUR_USER_ACCESS_TOKEN}
```

Only `execute` and `validate` can be used　in local, but not `publish`, `promote`, and `remove tag`.

## License
Code licensed under the BSD 3-Clause license. See LICENSE file for terms.

[build-image]: https://cd.screwdriver.cd/pipelines/408/badge
[build-url]: https://cd.screwdriver.cd/pipelines/408
[goreport-image]: https://goreportcard.com/badge/github.com/Screwdriver-cd/sd-cmd
[goreport-url]: https://goreportcard.com/report/github.com/Screwdriver-cd/sd-cmd
