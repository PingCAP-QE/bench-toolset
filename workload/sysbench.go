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
	sysbenchTpsRegex     = regexp.MustCompile(`\n\s+transactions:\s+([\d.]+)\s*\(([\d.]+)\s+per\s+sec\.\)`)
	sysbenchQpsRegex     = regexp.MustCompile(`\n\s+queries:\s+([\d.]+)\s*\(([\d.]+)\s+per\s+sec\.\)`)
	sysbenchMinLatRegex  = regexp.MustCompile(`\n\s+min:\s+([\d.]+)`)
	sysbenchAvgLatRegex  = regexp.MustCompile(`\n\s+avg:\s+([\d.]+)`)
	sysbenchMaxLatRegex  = regexp.MustCompile(`\n\s+max:\s+([\d.]+)`)
	sysbenchP99LatRegex  = regexp.MustCompile(`\n\s+99th percentile:\s+([\d.]+)`)
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

func (s *Sysbench) Records() ([]*Record, []*Record, error) {
	records, err := ParseSysbenchRecords(s.LogPath)
	if err != nil {
		return nil, nil, err
	}
	summaryRecord, err := ParseSysbenchSummaryReport(s.LogPath)
	return records, []*Record{summaryRecord}, err
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
			Payload:    make(map[string]interface{}),
		}
		percentage, err := strconv.ParseInt(string(matched[4]), 10, 64)
		switch percentage {
		case 95:
			p95Lat, err := strconv.ParseFloat(string(matched[5]), 64)
			if err != nil {
				return nil, errors.AddStack(err)
			}
			records[i].Payload["p95-lat"] = p95Lat
		case 99:
			p99Lat, err := strconv.ParseFloat(string(matched[5]), 64)
			if err != nil {
				return nil, errors.AddStack(err)
			}
			records[i].Payload["p99-lat"] = p99Lat
		}
	}
	return records, nil
}

func ParseSysbenchSummaryReport(logPath string) (*Record, error) {
	content, err := ioutil.ReadFile(logPath)
	if err != nil {
		return nil, err
	}
	summaryRecord := new(Record)
	summaryRecord.Type = "summary"
	tps := sysbenchTpsRegex.FindAllSubmatch(content, -1)
	if len(tps) == 1 {
		summaryRecord.Payload["tps"] = string(tps[0][2])
	}
	qps := sysbenchQpsRegex.FindAllSubmatch(content, -1)
	if len(qps) == 1 {
		summaryRecord.Payload["qps"] = string(qps[0][2])
	}
	minLat := sysbenchMinLatRegex.FindAllSubmatch(content, -1)
	if len(minLat) == 1 {
		summaryRecord.Payload["minLat"] = string(minLat[0][1])
	}
	avgLat := sysbenchAvgLatRegex.FindAllSubmatch(content, -1)
	if len(avgLat) == 1 {
		summaryRecord.Payload["avgLat"] = string(avgLat[0][1])
	}
	maxLat := sysbenchMaxLatRegex.FindAllSubmatch(content, -1)
	if len(maxLat) == 1 {
		summaryRecord.Payload["maxLat"] = string(maxLat[0][1])
	}
	p99Lat := sysbenchP99LatRegex.FindAllSubmatch(content, -1)
	if len(p99Lat) == 1 {
		summaryRecord.Payload["p99Lat"] = string(p99Lat[0][1])
	}
	return summaryRecord, nil
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
