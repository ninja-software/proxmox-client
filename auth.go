package proxmox

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// AuthTicketResponse is the response from the Proxmox API after a successful auth
type AuthTicketResponse struct {
	Data struct {
		Ticket              string `json:"ticket"`
		Username            string `json:"username"`
		CSRFPreventionToken string `json:"CSRFPreventionToken"`
	} `json:"data"`
}

// AuthResponse is the response from the Proxmox API container the CSRF token and ticket for authenticated requests
type AuthResponse struct {
	CSRFToken string `json:"csrfToken"`
	Ticket    string `json:"ticket"`
}

// SignIn will signin the user and update the embedded client with the ticket
func (c *Client) SignIn() error {
	log.WithFields(logrus.Fields{
		"username": c.username,
		"password": c.password,
	}).Debugln("Signing into Proxmox")

	u, err := url.Parse(c.host + "/api2/json/access/ticket")
	if err != nil {
		return errors.Wrap(err, "Could not parse URL")
	}

	resp, err := c.client.PostForm(u.String(), url.Values{
		"username": []string{c.username},
		"password": []string{c.password},
	})
	if err != nil {
		return errors.Wrap(err, "Could not POST form to proxmox ticket endpoint")
	}

	authResponse := &AuthTicketResponse{}

	MustDecodeJSON(resp.Body, authResponse)
	c.CSRFToken = authResponse.Data.CSRFPreventionToken
	c.Ticket = authResponse.Data.Ticket

	proxmoxCookie := &http.Cookie{
		Name:   "PVEAuthCookie",
		Value:  c.Ticket,
		Domain: proxmoxDomain,
		Path:   "/",
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Sprintf("Could not auth: %s", resp.Status)
		return errors.New(err)
	}

	c.client.Jar.SetCookies(u, []*http.Cookie{proxmoxCookie})
	log.Debugln("Successfully signed into Proxmox")
	return nil
}

// VerifyTicket confirms that the currently held ticket in the client is valid
func (c *Client) VerifyTicket() (bool, error) {
	log.Debugln("Checking Proxmox auth")
	u, err := url.Parse(c.host + "/api2/json/version")
	if err != nil {
		return false, errors.Wrap(err, "Could not parse URL")
	}

	resp, err := c.client.Get(u.String())
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Could not do auth check: %s", resp.Status)
	}

	return true, nil
}
