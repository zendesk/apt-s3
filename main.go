package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/zendesk/apt-s3/method"
)

var (
	downloadUri  = flag.String("download", "", "S3 URI for downloading a single file")
	downloadPath = flag.String("path", "", "Path to download file to")
	versionFlag  = flag.Bool("version", false, "Show version")
	Version      = "master"
)

func main() {
	ctx := context.Background()
	m := method.New(ctx)
	programName := os.Args[0]

	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s %s (Go version: %s)\n", programName, Version, runtime.Version())
		os.Exit(0)
		// Called outside of apt to download a file
	} else if *downloadUri != "" {
		filename, err := m.Downloader.DownloadFile(*downloadUri, *downloadPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("Downloaded %s\n", filename)
		os.Exit(0)
	} else {
		m.Start()
	}
}
