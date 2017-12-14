package proxmox

import (
	"fmt"
	"net/url"
)

// ISO is the response from the proxmox API for ISO content in a storage
type ISO struct {
	Content string `json:"content"`
	Volid   string `json:"volid"`
	Format  string `json:"format"`
	Size    int    `json:"size"`
}

// ISOListResponse is a list of ISOs from the Proxmox API
type ISOListResponse struct {
	Data []*ISO `json:"data"`
}

// ISOList returns a list of ISOs
func (c *Client) ISOList(node string) ([]string, error) {
	log.Debugln("Getting ISOs from Proxmox")
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

	u, err := url.Parse(fmt.Sprintf(c.host+"/api2/json/nodes/%s/storage/ISOs/content", node))
	proxmoxAPIResp, err := c.client.Get(u.String())
	if err != nil {
		return nil, err
	}
	dump(proxmoxAPIResp)
	proxmoxResponse := &ISOListResponse{}
	MustDecodeJSON(proxmoxAPIResp.Body, &proxmoxResponse)
	log.Println(proxmoxResponse)
	result := []string{}
	for _, iso := range proxmoxResponse.Data {
		result = append(result, iso.Volid)
	}
	return result, nil
}
