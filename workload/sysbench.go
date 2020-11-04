package workload

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

var (
	recordRegexp = regexp.MustCompile(`\[\s\d+s\s\]\sthds:\s(\d+)\stps:\s([\d\.]+)\sqps:\s([\d\.]+)\s[\(\)\w/:\s\d\.]+\slat\s\(ms,99%\):\s([\d\.]+)`)
)

type Sysbench struct {
	Host string
	User string
	Port uint64
	Db   string

	Tables    uint64
	TableSize uint64

	Name           string
	Threads        uint64
	Time           time.Duration
	ReportInterval time.Duration
	LogPath        string
}

func (s *Sysbench) Start() error {
	args := []string{
		s.Name,
		"--mysql-host=" + s.Host,
		"--mysql-user=" + s.User,
		"--mysql-port=" + fmt.Sprintf("%d", s.Port),
		"--tables=" + fmt.Sprintf("%d", s.Tables),
		"--table-size=" + fmt.Sprintf("%d", s.TableSize),
		"--threads=" + fmt.Sprintf("%d", s.Threads),
		"--time" + fmt.Sprintf("%1.0f", s.Time.Seconds()),
		"--report-interval" + fmt.Sprintf("%1.0f", s.Time.Seconds()),
		"--percentile=99%",
	}
	cmd := exec.Command("sysbench", args...)
	if len(s.LogPath) > 0 {
		logFile, err := os.OpenFile(s.LogPath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}
	return cmd.Run()
}

func (s *Sysbench) Records() ([]*Record, error) {
	return s.parseLogFile()
}

func (s *Sysbench) parseLogFile() ([]*Record, error) {
	content, err := ioutil.ReadFile(s.LogPath)
	if err != nil {
		return nil, err
	}
	matchedRecords := recordRegexp.FindAllSubmatch(content, -1)
	records := make([]*Record, len(matchedRecords))
	for i, matched := range matchedRecords {
		threads, err := strconv.ParseFloat(string(matched[1]), 64)
		if err != nil {
			return nil, err
		}
		tps, err := strconv.ParseFloat(string(matched[2]), 64)
		if err != nil {
			return nil, err
		}
		p99Lat, err := strconv.ParseFloat(string(matched[3]), 64)
		if err != nil {
			return nil, err
		}
		avgLat := 1000 / tps * threads
		records[i] = &Record{
			Count:   tps,
			Latency: &Latency{Avg: avgLat, P99: p99Lat},
			Time:    time.Second,
		}
	}

	return records, nil
}
