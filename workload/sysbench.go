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
	sysbenchRecordRegexp = regexp.MustCompile(`\[\s(\d+s)\s\]\sthds:\s(\d+)\stps:\s([\d\.]+)\sqps:\s[\d\.]+\s[\(\)\w/:\s\d\.]+\slat\s\(ms,(\d+)%\):\s([\d\.]+)`)
)

type Sysbench struct {
	Host string
	User string
	Port uint64
	Db   string

	Tables    uint64
	TableSize uint64

	Name    string
	Threads uint64
	Time    time.Duration
	LogPath string
}

func (s *Sysbench) Prepare() error {
	args := s.buildArgs()
	args = append(args, "prepare")
	cmd := exec.Command("sysbench", args...)
	return errors.Wrapf(cmd.Run(), "Sysbench prepare failed: args %v", cmd.Args)
}

func (s *Sysbench) Start() error {
	args := s.buildArgs()
	args = append(args, "run")
	cmd := exec.Command("sysbench", args...)
	if len(s.LogPath) > 0 {
		logFile, err := os.OpenFile(s.LogPath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}
	return errors.Wrapf(cmd.Run(), "Sysbench run failed: args %v", cmd.Args)
}

func (s *Sysbench) Records() ([]*Record, error) {
	return ParseSysbenchRecords(s.LogPath)
}

func ParseSysbenchRecords(logPath string) ([]*Record, error) {
	content, err := ioutil.ReadFile(logPath)
	if err != nil {
		return nil, err
	}
	matchedRecords := sysbenchRecordRegexp.FindAllSubmatch(content, -1)
	records := make([]*Record, len(matchedRecords))
	for i, matched := range matchedRecords {
		threads, err := strconv.ParseFloat(string(matched[2]), 64)
		if err != nil {
			return nil, errors.AddStack(err)
		}
		tps, err := strconv.ParseFloat(string(matched[3]), 64)
		if err != nil {
			return nil, errors.AddStack(err)
		}
		avgLat := 1000 / tps * threads
		records[i] = &Record{
			Tag:        string(matched[1]),
			Count:      tps,
			AvgLatInMs: avgLat,
		}
		percentage, err := strconv.ParseInt(string(matched[4]), 10, 64)
		switch percentage {
		case 95:
			p95Lat, err := strconv.ParseFloat(string(matched[5]), 64)
			if err != nil {
				return nil, errors.AddStack(err)
			}
			records[i].P95LatInMs = p95Lat
		case 99:
			p99Lat, err := strconv.ParseFloat(string(matched[5]), 64)
			if err != nil {
				return nil, errors.AddStack(err)
			}
			records[i].P99LatInMs = p99Lat
		}
	}

	return records, nil
}

func (s *Sysbench) buildArgs() []string {
	return []string{
		s.Name,
		"--mysql-host=" + s.Host,
		"--mysql-user=" + s.User,
		"--mysql-db=" + s.Db,
		"--mysql-port=" + fmt.Sprintf("%d", s.Port),
		"--tables=" + fmt.Sprintf("%d", s.Tables),
		"--table-size=" + fmt.Sprintf("%d", s.TableSize),
		"--threads=" + fmt.Sprintf("%d", s.Threads),
		"--time=" + fmt.Sprintf("%1.0f", s.Time.Seconds()),
		"--report-interval=1",
		"--percentile=99",
	}
}
