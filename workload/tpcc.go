package workload

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/pingcap/errors"
)

var (
	tpccRecordRegexp = regexp.MustCompile(`\[Current\]\s([\w]+)\s-\sTakes\(s\):\s([\d\.]+),\sCount:\s(\d+),\sTPM:\s[\d\.]+,\sSum\(ms\):\s[\d\.]+,\sAvg\(ms\):\s([\d\.]+),\s90th\(ms\):\s([\d\.]+),\s99th\(ms\):\s([\d\.]+),\s99\.9th\(ms\):\s([\d\.]+)`)
)

type Tpcc struct {
	WareHouses uint64
	Db         string
	Host       string
	Port       uint64
	User       string
	Threads    uint64
	Time       time.Duration
	LogPath    string
}

func (t *Tpcc) Prepare() error {
	args := t.buildArgs()
	args = append([]string{"tpcc", "prepare"}, args...)
	cmd := exec.Command("go-tpc", args...)
	return errors.Wrapf(cmd.Run(), "Tpcc prepare failed: args %v", cmd.Args)
}

func (t *Tpcc) Start() error {
	args := t.buildArgs()
	args = append([]string{"tpcc", "run"}, args...)
	cmd := exec.Command("go-tpc", args...)
	if len(t.LogPath) > 0 {
		logFile, err := os.OpenFile(t.LogPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}
	return errors.Wrapf(cmd.Run(), "Tpcc run failed: args %v", cmd.Args)
}

func (t *Tpcc) Records() ([]*Record, error) {
	return ParseTpccRecords(t.LogPath)
}

func ParseTpccRecords(logPath string) ([]*Record, error) {
	content, err := ioutil.ReadFile(logPath)
	if err != nil {
		return nil, err
	}
	matchedRecords := tpccRecordRegexp.FindAllSubmatch(content, -1)
	records := make([]*Record, len(matchedRecords))
	for i, matched := range matchedRecords {
		count, err := strconv.ParseFloat(string(matched[3]), 64)
		if err != nil {
			return nil, errors.AddStack(err)
		}
		avgLat, err := strconv.ParseFloat(string(matched[4]), 64)
		if err != nil {
			return nil, errors.AddStack(err)
		}
		p99Lat, err := strconv.ParseFloat(string(matched[6]), 64)
		if err != nil {
			return nil, errors.AddStack(err)
		}
		records[i] = &Record{
			Type:       string(matched[1]),
			Count:      count,
			AvgLatInMs: avgLat,
			P99LatInMs: p99Lat,
		}
	}

	return records, nil
}

func (t *Tpcc) buildArgs() []string {
	return []string{
		"--warehouses=" + fmt.Sprintf("%d", t.WareHouses),
		"--host=" + t.Host,
		"--port=" + fmt.Sprintf("%d", t.Port),
		"--user=" + t.User,
		"--time=" + t.Time.String(),
		"--db=" + t.Db,
		"--interval=1s",
	}
}
