package workload

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/pingcap/errors"
)

var (
	ycsbRecordRegexp = regexp.MustCompile(`(\w+)\s+-\sTakes\(s\):\s([\d\.]+),\sCount:\s\d+,\sOPS:\s([\d\.]+),\sAvg\(us\):\s(\d+),\sMin\(us\):\s\d+,\sMax\(us\):\s\d+,\s99th\(us\):\s(\d+),\s99\.9th\(us\):\s(\d+),\s99\.99th\(us\):\s(\d+)`)
)

type YcsbTarget interface {
	Args() []string
}

type YcsbTidbTarget struct {
	Host string
	Port uint64
}

func (y *YcsbTidbTarget) Args() []string {
	return []string{
		"mysql",
		"-p mysql.host=" + y.Host,
		"-p mysql.port=" + fmt.Sprintf("%d", y.Port),
	}
}

type YcsbTikvTarget struct {
	Pd string
}

func (y *YcsbTikvTarget) Args() []string {
	return []string{
		"tikv",
		"-p tikv.pd=" + y.Pd,
	}
}

type Ycsb struct {
	Workload string
	Target   YcsbTarget

	Threads        uint64
	OperationCount uint64
	RecordCount    uint64
	LogPath        string
}

func (y *Ycsb) Prepare() error {
	args := y.buildArgs()
	args = append([]string{"load"}, args...)
	cmd := exec.Command("go-ycsb", args...)
	return errors.Wrapf(cmd.Run(), "Ycsb load failed: args %v", cmd.Args)
}

func (y *Ycsb) Start() error {
	args := y.buildArgs()
	args = append([]string{"run"}, args...)
	cmd := exec.Command("go-ycsb", args...)
	if len(y.LogPath) > 0 {
		logFile, err := os.OpenFile(y.LogPath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}
	return errors.Wrapf(cmd.Run(), "Ycsb run failed: args %v", cmd.Args)
}

func (y *Ycsb) Records() ([]*Record, error) {
	content, err := ioutil.ReadFile(y.LogPath)
	if err != nil {
		return nil, err
	}
	matchedRecords := ycsbRecordRegexp.FindAllSubmatch(content, -1)
	records := make([]*Record, len(matchedRecords))
	for i, matched := range matchedRecords {
		ops, err := strconv.ParseFloat(string(matched[2]), 64)
		if err != nil {
			return nil, errors.AddStack(err)
		}
		avgLatInUs, err := strconv.ParseFloat(string(matched[3]), 64)
		if err != nil {
			return nil, errors.AddStack(err)
		}
		p99LatInUs, err := strconv.ParseFloat(string(matched[4]), 64)
		if err != nil {
			return nil, errors.AddStack(err)
		}
		records[i] = &Record{
			Type:       string(matched[0]),
			Count:      ops,
			AvgLatInMs: avgLatInUs / 1000,
			P99LatInMs: p99LatInUs / 1000,
		}
	}

	return records, nil
}

func (y *Ycsb) buildArgs() []string {
	targetArgs := y.Target.Args()
	return append(targetArgs, []string{
		"-P " + y.Workload,
		"-p threads=" + fmt.Sprintf("%d", y.Threads),
		"-p operationcount=" + fmt.Sprintf("%d", y.OperationCount),
		"-p recordcount=" + fmt.Sprintf("%d", y.RecordCount),
		"--interval=1",
	}...)
}
