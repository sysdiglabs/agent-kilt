# Handler

The `handler` is designed to be deployed as an AWS lambda to apply transformations to a task definition.

## Build

The `handler` uses `goreleaser` for builds and releases. The defintion can be found in `runtimes/cloudformation/.goreleaser.yml`.

*Note* to build the FIPS variant, you need to have [zig](https://github.com/ziglang/zig/wiki/Install-Zig-from-a-Package-Manager) installed.

```bash
$ cd runtimes/cloudformation
$ GORELEASER_CURRENT_TAG=0.0.1 goreleaser build --skip=validate --clean
...
$ ls -R dist/
dist/:
artifacts.json  config.yaml  handler  metadata.json

dist/handler:
handler-fips-linux-amd64  handler-fips-linux-arm64  handler-linux-amd64  handler-linux-arm64
```

## Test

```bash
$ cd runtimes/cloudformation
$ go test -race -mod=readonly ./...
```
