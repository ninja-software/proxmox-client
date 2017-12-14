package proxmox

import (
	"fmt"
	"net/url"
)

// Template is the values needed to create a container based on a template
type Template struct {
	Content string `json:"content"`
	Volid   string `json:"volid"`
	Format  string `json:"format"`
	Size    int    `json:"size"`
}

// TemplateListResponse is a list of Templates
type TemplateListResponse struct {
	Data []*Template `json:"data"`
}

// TemplateList returns a list of templates
func (c *Client) TemplateList(node string) ([]*Template, error) {
	log.Debugln("Getting templates from Proxmox")
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

	u, err := url.Parse(fmt.Sprintf(c.host+"/api2/json/nodes/%s/storage/templates/content", node))
	proxmoxAPIResp, err := c.client.Get(u.String())
	if err != nil {
		return nil, err
	}
	proxmoxResponse := &TemplateListResponse{}
	MustDecodeJSON(proxmoxAPIResp.Body, &proxmoxResponse)

	return proxmoxResponse.Data, nil
}
