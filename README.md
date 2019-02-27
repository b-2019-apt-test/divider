# Divider

Divider is a tiny CLI tool intended for passing aptitude test. Divider reads a stream of JSON-encoded jobs from file, processes them with call to external library and writes results to another file in CSV format.

## Getting started

Install with `go install`:
```
$ go get -u github.com/b-2019-apt-test/divider
$ go install github.com/b-2019-apt-test/divider/cmd/divider
```

### Usage

The only required argument is the path to the file with jobs:
```
$ divider -i jobs.json
```

By default, divider writes results to **divider.csv** file at the same directory level. You can specify alternative path:
```
$ divider -i jobs.json -o results.csv
```

You can specify division method with `-m` flag:
```
$ divider -i jobs.json -m cgo
```

Available methods: `go`, `cgo`, `syscall` (default). Option `-z` has precedence over the `-m` flag: if both defined, `go` method will be used.

Run divider without arguments to see the full usage info.

## Dependencies

Divider depends on external library **math.dll**: it's assumed that the library is already installed on the system. Although, you can run divider with its own division implementation using `-z` option:
```
$ divider -i jobs.json -z
```

For `cgo` method to work, the library must be installed on the system as **magic.dll**.

## Benchmarks

Benchmarks for `calldiv` and `cgodiv` demonstrate almost linear scaling:
```
$ go test github.com/b-2019-apt-test/divider/pkg/div/... -v -run none -bench=Parallel$ -benchmem -benchtime=5s -cpu 1,2,4,8,16,32,64,128
```

## Build and test

**pkg-config** is used for `cgo` dependencies management. You can find sample config under the **build** directory of the repository.
