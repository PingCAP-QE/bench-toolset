package cluster

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

const (
	resourceRequestApi = "/api/cluster/%v"
	reportApi          = "/api/cluster/workload/%v/result"
)

var dialClient = &http.Client{}

func doResourceRequest(addr string, name string, content []byte) (uint64, error) {
	url := addr + fmt.Sprint(resourceRequestApi, name)
	resp, err := dialClient.Post(url, "application/json", bytes.NewBuffer(content))
	if err != nil {
		return 0, err
	}
	var data []byte
	_, err = resp.Body.Read(data)
	if err != nil {
		return 0, err
	}
	var respFields map[string]string
	err = json.Unmarshal(data, &respFields)
	if err != nil {
		return 0, err
	}

	return strconv.ParseUint(respFields["cluster_request_id"], 10, 64)
}

func doResourceStatusRequest(addr string, id uint64) (Status, error) {
	url := addr + fmt.Sprint(resourceRequestApi, id)
	resp, err := dialClient.Get(url)
	if err != nil {
		return StatusOther, err
	}
	var data []byte
	_, err = resp.Body.Read(data)
	if err != nil {
		return StatusOther, err
	}
	var respFields map[string]string
	err = json.Unmarshal(data, &respFields)
	if err != nil {
		return StatusOther, err
	}
	switch respFields["status"] {
	case "READY":
		return StatusReady, nil
	case "DONE":
		return StatusDone, nil
	}
	return StatusOther, nil
}
