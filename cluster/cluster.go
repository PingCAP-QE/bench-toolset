package cluster

import "encoding/json"

type Status uint64

const (
	StatusReady Status = iota
	StatusDone
	StatusOther
)

// Topology defines the topology of the cluster.
// Each field represents resource ids of corresponding component.
type Topology struct {
	Pd         []uint64
	Tidb       []uint64
	Tikv       []uint64
	Prometheus uint64
	Grafana    uint64
	Path       string
}

func (t Topology) MarshalJSON() ([]byte, error) {
	type clusterItem struct {
		Id         uint64 `json:"rri_item_id"`
		Component  string `json:"component"`
		DeployPath string `json:"deploy_path"`
	}
	items := make([]*clusterItem, 0)
	for _, id := range t.Pd {
		items = append(items, &clusterItem{Id: id, Component: "pd", DeployPath: t.Path})
	}
	for _, id := range t.Tidb {
		items = append(items, &clusterItem{Id: id, Component: "tidb", DeployPath: t.Path})
	}
	for _, id := range t.Tikv {
		items = append(items, &clusterItem{Id: id, Component: "tikv", DeployPath: t.Path})
	}
	items = append(items, &clusterItem{Id: t.Prometheus, Component: "prometheus", DeployPath: t.Path})
	items = append(items, &clusterItem{Id: t.Grafana, Component: "grafana", DeployPath: t.Path})

	return json.Marshal(items)
}

type Workload struct {
	Rid         uint64   `json:"rri_item_id"`
	DockerImage string   `json:"docker_image"`
	RestorePath string   `json:"restore_path"`
	Cmd         string   `json:"cmd"`
	Args        []string `json:"args"`
}

type Meta struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Cluster struct {
	id      uint64
	apiAddr string

	Meta     *Meta     `json:"cluster_request"`
	Topology *Topology `json:"cluster_request_topologies"`
	Workload *Workload `json:"cluster_workload"`
}

func NewCluster(name string, apiAddr string, version string, topology *Topology, workload *Workload) *Cluster {
	if len(topology.Path) == 0 {
		topology.Path = "/" + name
	}
	return &Cluster{
		id: 0,
		Meta: &Meta{
			Name:    name,
			Version: version,
		},
		apiAddr:  apiAddr,
		Topology: topology,
		Workload: workload,
	}
}

func (c *Cluster) Run() (uint64, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return 0, err
	}
	id, err := doResourceRequest(c.apiAddr, c.Meta.Name, data)
	c.id = id
	return id, err
}

func (c *Cluster) Status() (Status, error) {
	return doResourceStatusRequest(c.apiAddr, c.id)
}
