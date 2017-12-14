package proxmox

import (
	"fmt"
	"math"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// PickNode returns node with the least provisioned memory
func (c *Client) PickNode() (string, error) {
	resources, err := c.ResourceList()
	if err != nil {
		return "", err
	}
	nodes := resources.Nodes()
	if len(nodes) < 1 {
		return "", errors.New("no nodes found")
	}

	containers := resources.Containers()

	result := nodes[0].Node
	mostNonProvisionedMemory := math.MinInt32
	for _, node := range nodes {
		nonProvisionedMemory := node.Mem

		// Subtract total provisioned memory for this node
		for _, container := range containers {
			if container.Node == node.Node {
				nonProvisionedMemory -= container.Mem
			}
		}

		// Use this node if it has more non provisioned memory than the previous record
		if nonProvisionedMemory > mostNonProvisionedMemory {
			result = node.Node
			mostNonProvisionedMemory = nonProvisionedMemory
		}
	}
	return result, nil
}

// NodeStatus is the response from the Proxmox API
type NodeStatus struct {
	CPU     int `json:"cpu"`
	Cpuinfo struct {
		Cpus    int    `json:"cpus"`
		Hvm     int    `json:"hvm"`
		Mhz     string `json:"mhz"`
		Model   string `json:"model"`
		Sockets int    `json:"sockets"`
		UserHz  int    `json:"user_hz"`
	} `json:"cpuinfo"`
	Idle int `json:"idle"`
	Ksm  struct {
		Shared int `json:"shared"`
	} `json:"ksm"`
	Kversion string   `json:"kversion"`
	Loadavg  []string `json:"loadavg"`
	Memory   struct {
		Free  int64 `json:"free"`
		Total int64 `json:"total"`
		Used  int   `json:"used"`
	} `json:"memory"`
	Pveversion string `json:"pveversion"`
	Rootfs     struct {
		Avail int64 `json:"avail"`
		Free  int64 `json:"free"`
		Total int64 `json:"total"`
		Used  int64 `json:"used"`
	} `json:"rootfs"`
	Swap struct {
		Free  int64 `json:"free"`
		Total int64 `json:"total"`
		Used  int   `json:"used"`
	} `json:"swap"`
	Uptime int `json:"uptime"`
	Wait   int `json:"wait"`
}

// NodeStatusResponse is the response from the Proxmox API for a node's status
type NodeStatusResponse struct {
	Data *NodeStatus `json:"data"`
}

// NodeStatus returns the Node's RAM, CPU and storage
func (c *Client) NodeStatus(node string) (*NodeStatus, error) {
	log.Debugln("Getting node stats")

	authed, err := c.VerifyTicket()
	if err != nil {
		return nil, err
	}

	if !authed {
		err = c.SignIn()
		if err != nil {
			return nil, err
		}
	}

	u, err := url.Parse(fmt.Sprintf("%s/api2/json/nodes/%s/status", c.host, node))
	if err != nil {
		return nil, err
	}

	proxmoxAPIResp, err := c.client.Get(u.String())
	if proxmoxAPIResp.StatusCode != http.StatusOK {
		err := fmt.Sprintf("Could not get node status: %s", proxmoxAPIResp.Status)
		return nil, errors.New(err)
	}

	proxmoxResp := &NodeStatusResponse{}

	MustDecodeJSON(proxmoxAPIResp.Body, proxmoxResp)
	return proxmoxResp.Data, nil
}
