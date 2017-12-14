package proxmox

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// ContainerCreateRequest is a request to the Proxmox API to create a container
type ContainerCreateRequest struct {
	MAC             string
	Template        *ParsedTemplate
	Node            string
	VMID            int
	CPUCores        int
	Memory          int
	StorageCapacity int
	StorageID       string
	SSHPublicKey    string
	Password        string
	HostName        string
	IPAddress       string
}

// ParsedTemplate is the result of magical regex on a filepath to get template info
type ParsedTemplate struct {
	OS         string
	OSVersion  string
	Name       string
	OSVersion2 string
	Arch       string
	Extension  string
}

// String returns a string version of the template struct
func (pt *ParsedTemplate) String() string {
	return fmt.Sprintf("%s-%s-%s_%s_%s%s", pt.OS, pt.OSVersion, pt.Name, pt.OSVersion2, pt.Arch, pt.Extension)
}

// ContainerConfigRequest is a request to the Proxmox API for a VM's configuration
type ContainerConfigRequest struct {
	Node string
	VMID int
}

// ContainerConfig is the response from the Proxmox API for a VM's configuration
type ContainerConfig struct {
	Memory      int    `json:"memory"`
	Cpulimit    string `json:"cpulimit"`
	Digest      string `json:"digest"`
	Cores       int    `json:"cores"`
	Ostype      string `json:"ostype"`
	Rootfs      string `json:"rootfs"`
	Hostname    string `json:"hostname"`
	Arch        string `json:"arch"`
	Description string `json:"description"`
	Swap        int    `json:"swap"`
	Net0        string `json:"net0"`
}

// ContainerConfigResponse is the response from the Proxmox API for a VM's configuration
type ContainerConfigResponse struct {
	Data *ContainerConfig `json:"data"`
}

// ContainerVMStatusRequest is the request for the Proxmox API to modify a container or VM's status
type ContainerVMStatusRequest struct {
	Node      string
	VMID      int
	StorageID string
}

// ContainerConfig returns the container config
func (c *Client) ContainerConfig(params *ContainerConfigRequest) (*ContainerConfig, error) {
	log.WithFields(logrus.Fields{
		"node": params.Node,
		"vmid": params.VMID,
	}).Debugln("Getting container config")

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

	u, err := url.Parse(fmt.Sprintf("%s/api2/json/nodes/%s/lxc/%d/config", c.host, params.Node, params.VMID))
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Get(u.String())
	if resp.StatusCode != http.StatusOK {
		err := fmt.Sprintf("Could not get container config: %s", resp.Status)
		return nil, errors.New(err)
	}
	result := &ContainerConfigResponse{}
	MustDecodeJSON(resp.Body, result)
	return result.Data, nil
}

// ContainerCreate runs the Create action for the Proxmox containers
func (c *Client) ContainerCreate(params *ContainerCreateRequest) error {
	log.WithFields(logrus.Fields{
		"MAC":             params.MAC,
		"Template":        params.Template,
		"Node":            params.Node,
		"VMID":            params.VMID,
		"CPUCores":        params.CPUCores,
		"Memory":          params.Memory,
		"StorageCapacity": params.StorageCapacity,
		"StorageID":       params.StorageID,
		"SSHPublicKey":    params.SSHPublicKey,
		"Password":        params.Password,
		"HostName":        params.HostName,
		"IPAddress":       params.IPAddress,
	}).Debugln("Creating container")

	authed, err := c.VerifyTicket()
	if err != nil {
		return err
	}

	if !authed {
		err = c.SignIn()
		if err != nil {
			return err
		}
	}

	u, err := url.Parse(fmt.Sprintf("%s/api2/json/nodes/%s/lxc", c.host, params.Node))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return errors.Wrap(err, "Could not create request")
	}

	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(params); err != nil {
		log.Fatal(err)
	}

	q := req.URL.Query()
	q.Add("ostemplate", "templates:vztmpl/"+params.Template.String())
	q.Add("vmid", strconv.Itoa(params.VMID))
	q.Add("storage", params.StorageID)
	q.Add("swap", "512")
	q.Add("cores", strconv.Itoa(params.CPUCores))
	q.Add("rootfs", fmt.Sprintf("%d", params.StorageCapacity))
	q.Add("cpulimit", strconv.Itoa(params.CPUCores))
	q.Add("memory", strconv.Itoa(params.Memory))
	q.Add("hostname", params.HostName)
	q.Add("password", params.Password)
	q.Add("description", buf.String())
	q.Add("net0", fmt.Sprintf("name=eth0,bridge=vmbr3,hwaddr=%s,ip=dhcp,tag=10,type=veth", params.MAC))

	req.URL.RawQuery = q.Encode()
	req.Header.Set("CSRFPreventionToken", c.CSRFToken)

	proxmoxResp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "Could not execute request")
	}
	if proxmoxResp.StatusCode != http.StatusOK {
		dump(proxmoxResp)
		err := fmt.Sprintf("Could not create container: %s", proxmoxResp.Status)
		return errors.New(err)
	}
	return nil
}

// ContainerAdd runs the Add action for the Proxmox containers
func (c *Client) ContainerAdd(container *Resource) error { panic("Not implemented") }

// ContainerUpdate runs the Update action for the Proxmox containers
func (c *Client) ContainerUpdate(container *Resource) error { panic("Not implemented") }

// ContainerDelete runs the Delete action for the Proxmox containers
func (c *Client) ContainerDelete(node string, vmid int) error {
	log.WithFields(logrus.Fields{
		"node": node,
		"vmid": strconv.Itoa(vmid),
	}).Debugln("Deleting container")
	u, err := url.Parse(fmt.Sprintf("%s/api2/json/nodes/%s/lxc/%s", c.host, node, strconv.Itoa(vmid)))
	if err != nil {
		return err
	}

	authed, err := c.VerifyTicket()
	if err != nil {
		return err
	}

	if !authed {
		err = c.SignIn()
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return errors.Wrap(err, "Could not create request")
	}

	req.Header.Set("CSRFPreventionToken", c.CSRFToken)

	proxmoxResp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "Could not execute request")
	}

	if proxmoxResp.StatusCode != http.StatusOK {
		dump(proxmoxResp)
		err := fmt.Sprintf("Could not create container: %s", proxmoxResp.Status)
		return errors.New(err)
	}
	return nil
}

// ContainerStop will stop the container
func (c *Client) ContainerStop(params *ContainerVMStatusRequest) error {
	return c.vmStatusPOSTHelper("stop", params.Node, params.VMID, "lxc")
}

// ContainerStart will start the container
func (c *Client) ContainerStart(params *ContainerVMStatusRequest) error {
	return c.vmStatusPOSTHelper("start", params.Node, params.VMID, "lxc")
}

// ContainerShutdown will shutdown the container
func (c *Client) ContainerShutdown(params *ContainerVMStatusRequest) error {
	return c.vmStatusPOSTHelper("shutdown", params.Node, params.VMID, "lxc")
}

// ContainerResume will start the container
func (c *Client) ContainerResume(params *ContainerVMStatusRequest) error {
	return c.vmStatusPOSTHelper("resume", params.Node, params.VMID, "lxc")
}
