package main

import (
	"fmt"
	"os"
	"time"

	api "github.com/cryptix/goPshdlRest"
)

func downloadAllFiles(client api.Client) error {
	wspace, resp, err := client.Workspace.GetInfo()
	if err != nil {
		return err
	}

	// probe folder
	err = os.Mkdir(wspace.Id, 0700)
	if err != nil {
		return err
	}

	os.Chdir(wspace.Id)

	done := make(chan error)
	fileCount := len(wspace.Files)

	if fileCount == 0 {
		return nil
	}

	start := time.Now()

	for _, file := range wspace.Files {
		go func() {
			fmt.Println("Download files TODO")
			// f.DownloadFile(done)
		}()
	}

	for {
		select {
		case err := <-done:
			if err != nil {
				close(done)
				return fmt.Errorf("Could not load all files. %s", err)
			}

			fileCount -= 1
			if fileCount == 0 {
				os.Chdir("..")
				return nil
			}

		case <-time.After(3 * time.Second):
			fmt.Fprintf(os.Stderr, "Waiting.. %d files left. Duration: %s\n", fileCount, time.Since(start))
		}
	}
}
