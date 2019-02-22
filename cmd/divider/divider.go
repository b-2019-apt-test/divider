//+build windows

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/b-2019-apt-test/divider/internal/divider"
	"github.com/b-2019-apt-test/divider/internal/divider/extra"
	"github.com/b-2019-apt-test/divider/internal/divider/ordinary"
)

const (
	jobsUsage   = `path to the file with jobs`
	jobsDefault = ``

	resultsUsage   = `results file path`
	resultsDefault = "divider.csv"

	logUsage   = `log file path`
	logDefault = ``

	workersUsage        = `workers count`
	workersDefault uint = 256

	sizeUsage        = `worker pool buffer size`
	sizeDefault uint = 2048

	dontUseExtUsage   = `do not use math.dll`
	dontUseExtDefault = false
)

func main() {

	var (
		jobsFilePath    string
		resultsFilePath string
		logFilePath     string
		workers         uint
		size            uint
		dontUseExt      bool
	)

	flag.StringVar(&jobsFilePath, "i", jobsDefault, jobsUsage)
	flag.StringVar(&resultsFilePath, "o", resultsDefault, resultsUsage)
	flag.StringVar(&logFilePath, "log", logDefault, logUsage)
	flag.UintVar(&workers, "w", workersDefault, workersUsage)
	flag.UintVar(&size, "s", sizeDefault, sizeUsage)
	flag.BoolVar(&dontUseExt, "z", dontUseExtDefault, dontUseExtUsage)
	flag.Parse()

	if len(jobsFilePath) == 0 {
		fmt.Fprintln(os.Stderr, "Path to the jobs file not specified.")
		flag.PrintDefaults()
		os.Exit(1)
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
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer closeFile(reader)

	writer, err := os.Create(resultsFilePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer closeSyncFile(writer)

	var div divider.Divider
	if dontUseExt {
		div = ordinary.NewDivider()
	} else {
		div = extra.NewDivider()
	}

	proc := divider.NewJobProcessor().SetDivider(div).
		SetResultWriter(writer).
		SetJobReader(reader).
		SetWorkersCount(workers).
		SetPoolSize(size).
		SetLogger(logger)

	done := make(chan error, 1)
	go func() {
		done <- proc.Start()
	}()

	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)
	var ack bool

loop:
	for {
		select {
		case v := <-s:
			if ack {
				fmt.Println("Stopping...")
				continue
			}
			fmt.Printf("Signal received: %v. Stopping job processing...\n", v)
			proc.Stop()
			ack = true
		case err := <-done:
			if err != nil {
				logger.Fatal("Processing failed:", err)
			}
			fmt.Println("Complete.")
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
