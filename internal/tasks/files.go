package tasks

import (
	"encoding/json"
	"fmt"

	"github.com/cavaliercoder/grab"
	"github.com/cybertec-postgresql/pg_timetable/internal/pgengine"
)

type downloadOpts struct {
	WorkersNum int      `json:"workersnum"`
	FileUrls   []string `json:"fileurls"`
	DestPath   string   `json:"destpath"`
}

func taskDownloadFile(paramValues string) error {
	var opts downloadOpts
	if err := json.Unmarshal([]byte(paramValues), &opts); err != nil {
		return err
	}
	// create multiple download requests
	reqs := make([]*grab.Request, 0)
	for _, url := range opts.FileUrls {
		req, err := grab.NewRequest(opts.DestPath, url)
		if err != nil {
			return err
		}
		reqs = append(reqs, req)
	}

	// start downloads with workers, if < 0 => worker for each file
	client := grab.NewClient()
	respch := client.DoBatch(opts.WorkersNum, reqs...)

	// check each response
	var errstrings []string
	for resp := range respch {
		if err := resp.Err(); err != nil {
			errstrings = append(errstrings, err.Error())
		} else {
			pgengine.LogToDB("LOG", fmt.Sprintf("Downloaded %s to %s\n", resp.Request.URL(), resp.Filename))
		}
	}

	if len(errstrings) > 0 {
		return fmt.Errorf("download failed: %v", errstrings)
	}
	return nil
}
