package proxmox

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

// Resource is a VM, container or storage on proxmox
type Resource struct {
	CPU       float64 `json:"cpu,omitempty"`
	Disk      int     `json:"disk,omitempty"`
	Diskread  int     `json:"diskread,omitempty"`
	Diskwrite int     `json:"diskwrite,omitempty"`
	ID        string  `json:"id"`
	Maxcpu    int     `json:"maxcpu,omitempty"`
	Maxdisk   int64   `json:"maxdisk,omitempty"`
	Maxmem    int     `json:"maxmem,omitempty"`
	Mem       int     `json:"mem,omitempty"`
	Name      string  `json:"name,omitempty"`
	Netin     int     `json:"netin,omitempty"`
	Netout    int     `json:"netout,omitempty"`
	Node      string  `json:"node"`
	Status    string  `json:"status,omitempty"`
	Template  int     `json:"template,omitempty"`
	Type      string  `json:"type"`
	Uptime    int     `json:"uptime,omitempty"`
	Vmid      int     `json:"vmid,omitempty"`
	Level     string  `json:"level,omitempty"`
	Storage   string  `json:"storage,omitempty"`
}

// Resources is a list of resources from the Proxmox API
type Resources []*Resource

// GetNodeFromVMID will return a VM's node
func (pr Resources) GetNodeFromVMID(vmid int) (string, error) {
	for _, v := range pr {
		if v.Type == "lxc" || v.Type == "qemu" {
			if v.Vmid == vmid {
				return v.Node, nil
			}
		}
	}

	return "", errors.New("Could not find node for VMID: " + strconv.Itoa(vmid))
}

// Storages returns a slice of storages from the Proxmox API
func (pr Resources) Storages() []*Resource {
	result := []*Resource{}
	for _, v := range pr {
		if v.Type == "storage" {
			result = append(result, v)
		}
	}
	return result
}

// Templates returns a slice of templates from the Proxmox API
func (pr Resources) Templates() []*Resource {
	result := []*Resource{}
	for _, v := range pr {
		if v.Type == "template" {
			result = append(result, v)
		}
	}
	return result
}

// Containers returns a slice of containers from the Proxmox API
func (pr Resources) Containers() []*Resource {
	result := []*Resource{}
	for _, v := range pr {
		if v.Type == "lxc" {
			result = append(result, v)
		}
	}
	return result
}

// VMs returns a slice of VMs from the Proxmox API
func (pr Resources) VMs() []*Resource {
	result := []*Resource{}
	for _, v := range pr {
		if v.Type == "qemu" {
			result = append(result, v)
		}
	}
	return result
}

// Nodes returns a slice of nodes from the Proxmox API
func (pr Resources) Nodes() []*Resource {
	result := []*Resource{}
	for _, v := range pr {
		if v.Type == "node" {
			result = append(result, v)
		}
	}
	return result
}

// ResourceList runs the List action for the Proxmox resources
func (c *Client) ResourceList() (Resources, error) {
	log.Debugln("Getting resources from cluster")

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

	u, err := url.Parse(fmt.Sprintf("%s/api2/json/cluster/resources", c.host))
	if err != nil {
		return nil, err
	}
	proxmoxAPIResp, err := c.client.Get(u.String())
	if proxmoxAPIResp.StatusCode != http.StatusOK {
		err := fmt.Sprintf("Could not get node list: %s", proxmoxAPIResp.Status)
		return nil, errors.New(err)
	}
	proxmoxResp := &ResourcesResponse{}

	MustDecodeJSON(proxmoxAPIResp.Body, proxmoxResp)
	return Resources(proxmoxResp.Data), nil
}

// ResourcesResponse is a list of Resources from the Proxmox API
type ResourcesResponse struct {
	Data []*Resource `json:"data"`
}
