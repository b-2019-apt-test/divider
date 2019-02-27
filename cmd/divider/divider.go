//+build windows

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/b-2019-apt-test/divider/internal/divider"
	"github.com/b-2019-apt-test/divider/internal/divider/csvrep"
	"github.com/b-2019-apt-test/divider/internal/divider/jsonprov"

	"github.com/b-2019-apt-test/divider/pkg/div"
	"github.com/b-2019-apt-test/divider/pkg/div/calldiv"
	"github.com/b-2019-apt-test/divider/pkg/div/cgodiv"
	"github.com/b-2019-apt-test/divider/pkg/div/godiv"
)

var (
	jobsFilePath    string
	resultsFilePath string
	logFilePath     string
	workers         uint
	dontUseExt      bool
	method          string

	start = time.Now()
)

func main() {

	flag.StringVar(&jobsFilePath, "i", "", "path to the file with jobs")
	flag.StringVar(&resultsFilePath, "o", "divider.csv", "results file path")
	flag.StringVar(&logFilePath, "log", "", "log file path")
	flag.UintVar(&workers, "w", uint(runtime.NumCPU()*1024), "workers count")
	flag.BoolVar(&dontUseExt, "z", false, "do not use math.dll (depricated)")
	flag.StringVar(&method, "m", "syscall", "division method: go, cgo, syscall")
	flag.Parse()

	if len(flag.Args()) != 0 {
		exitWithUsage("Unparsed args:", flag.Args())
	}

	if len(jobsFilePath) == 0 {
		exitWithUsage("Path to the jobs file not specified.")
	}

	var (
		err error
		d   div.Divider
	)

	if dontUseExt {
		d = godiv.Divider
	} else if d = parseDivider(method); d == nil {
		exitWithUsage("Divider method unknown:", method)
	}

	logger := log.New(os.Stderr, "", log.LstdFlags)
	if len(logFilePath) != 0 {
		logFile, err := os.OpenFile(logFilePath,
			os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			logger.SetOutput(logFile)
			defer closeSyncFile(logFile)
		}
	}

	reader, err := os.OpenFile(jobsFilePath, os.O_RDONLY, 0400)
	exitOn(err)
	defer closeFile(reader)
	provider, err := jsonprov.New(reader)
	exitOn(err)

	writer, err := os.Create(resultsFilePath)
	exitOn(err)
	defer closeSyncFile(writer)
	reporter, err := csvrep.New(writer)
	exitOn(err)

	proc := divider.NewJobProcessor().
		SetJobProvider(provider).
		SetResultReporter(reporter).
		SetWorkersCount(workers).
		SetLogger(logger).
		SetDivider(d)

	done := make(chan error, 1)
	go func() {
		done <- proc.Start()
	}()

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	var ack bool

loop:
	for {
		select {
		case v := <-s:
			if ack {
				logger.Println("Stopping...")
				continue
			}

			logger.Printf("Signal received: %v. Stopping job processing...\n", v)
			proc.Stop()
			ack = true

		case err := <-done:
			if err != nil {
				logger.Fatal("Processing failed: ", err)
			}

			dur := time.Since(start)
			rate := float64(proc.Processed()) / dur.Seconds()

			logger.Println("Complete.")
			logger.Println("Processed jobs:", proc.Processed())
			logger.Printf("Time taken: %v, Avg. rate: %v\n", dur, uint64(rate))

			break loop
		}
	}
}

func closeSyncFile(f *os.File) {
	syncFile(f)
	closeFile(f)
}

func closeFile(f *os.File) {
	if err := f.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to close file %s: %v", f.Name(), err)
	}
}

func syncFile(f *os.File) {
	if err := f.Sync(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to sync file %s: %v", f.Name(), err)
	}
}

func exitOn(err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

func exitWithUsage(msg ...interface{}) {
	fmt.Fprintln(os.Stderr, msg...)
	flag.PrintDefaults()
	os.Exit(1)
}

func parseDivider(s string) div.Divider {
	var d div.Divider
	switch s {
	case "":
		fallthrough
	case "syscall":
		d = calldiv.Divider
	case "cgo":
		d = cgodiv.Divider
	case "go":
		d = godiv.Divider
	}
	return d
}
