package proxmox

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/pkg/errors"
)

// StorageCreate will create a storage for a VM
func (c *Client) StorageCreate(vmid, size string) error {
	log.Debugln("Creating storage in Proxmox")

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

	host := "example"
	port := "example"
	node := "example"
	storage := "example"
	u, err := url.Parse(fmt.Sprintf("https://%s:%s/api2/json/nodes/%s/storage/%s/content", host, port, node, storage))
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return errors.Wrap(err, "Could not create request")
	}

	q := req.URL.Query()
	q.Add("filename", fmt.Sprintf("vm-%s-disk-1", vmid))
	q.Add("format", "raw")
	q.Add("size", size+"G")
	q.Add("vmid", vmid)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("CSRFPreventionToken", c.CSRFToken)

	proxmoxResp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "Could not execute request")
	}
	if proxmoxResp.StatusCode != http.StatusOK {
		d, _ := httputil.DumpResponse(proxmoxResp, true)
		log.Errorln(string(d))
		err := fmt.Sprintf("Could not create container: %s", proxmoxResp.Status)
		return errors.New(err)
	}
	return nil
}
