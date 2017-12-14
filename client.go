// Package proxmox implements the VMProvider and ContainerProvider methods
// It makes requests to the Proxmox VE on behalf of Raijin
package proxmox

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"

	"github.com/blockninja/proxmox-client/logger"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// Client contains the state for the proxmox client
type Client struct {
	CSRFToken string
	Ticket    string
	host      string
	client    *http.Client
	username  string
	password  string
}

// New returns a new Proxmox client
func New(host, username, password string) (*Client, error) {
	log = logger.Get()
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create cookie jar")
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	client.Jar = jar
	result := &Client{
		username: username,
		password: password,
		host:     host,
		client:   client,
	}

	err = result.SignIn()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// NextID returns the next available VMID
func (c *Client) NextID() (int, error) {
	authed, err := c.VerifyTicket()
	if err != nil {
		return 0, err
	}

	if !authed {
		err = c.SignIn()
		if err != nil {
			return 0, err
		}
	}

	u, err := url.Parse(fmt.Sprintf("%s/api2/json/cluster/nextid", c.host))
	if err != nil {
		return 0, err
	}
	log.Debugln("Getting next available ID")

	proxmoxAPIResp, err := c.client.Get(u.String())
	if err != nil {
		return 0, err
	}
	if proxmoxAPIResp.StatusCode != http.StatusOK {
		err := fmt.Sprintf("Could not get next ID: %s", proxmoxAPIResp.Status)
		return 0, errors.New(err)
	}
	result := &NextIDResponse{}
	MustDecodeJSON(proxmoxAPIResp.Body, result)
	return strconv.Atoi(result.Data)
}
