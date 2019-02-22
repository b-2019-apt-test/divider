# Divider

Divider is a tiny CLI tool intended for passing aptitude test. Divider reads stream of JSON-encoded jobs from file, processes them with call to `math.dll` and writes results to another file in CSV format.

## Getting started

Install with `go install`:
```
$ go get github.com/b-2019-apt-test/divider/...
$ go install github.com/b-2019-apt-test/divider/cmd/divider
```

### Usage

The only required argument is the path to the file with jobs:
```
$ divider -i jobs.json
```

By default, divider writes results to `divider.csv` file at the same directory level. You can specify alternative path for the results:
```
$ divider -i jobs.json -o results.csv
```

Run divider without arguments to see full usage info.

## Dependencies

Divider depends on `math.dll`: it's assumed that the library is already installed on the system. Although, you can run divider with its own implementation of the division using `-z` option:
```
$ divider -i jobs.json -z
```
