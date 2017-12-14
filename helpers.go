package proxmox

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"

	"github.com/blockninja/ninjarouter"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

func (c *Client) vmStatusPOSTHelper(action string, node string, containerID int, vmType string) error {
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

	u, err := url.Parse(fmt.Sprintf("%s/api2/json/nodes/%s/%s/%d/status/%s", c.host, node, vmType, containerID, action))
	if err != nil {
		return errors.Wrap(err, "Could not parse URL")
	}

	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return errors.Wrap(err, "Could not create request")
	}
	req.Header.Set("CSRFPreventionToken", c.CSRFToken)

	proxmoxResp, err := c.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "Could not execute request")
	}
	if proxmoxResp.StatusCode != http.StatusOK {
		err := fmt.Sprintf("Could not %s container: %s", action, proxmoxResp.Status)
		return errors.New(err)
	}
	return nil
}

func dump(resp *http.Response) {
	d, _ := httputil.DumpResponse(resp, true)
	log.Debugln(string(d))
}

// Error is a struct with which we can marshal into JSON for a HTTP response
type Error struct {
	Message string
}

// MustReadAll will read the response body as a string
func MustReadAll(r io.Reader) string {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return string(buf)
}

// ErrorJSON returns a string for use in http.Error
func ErrorJSON(err error) string {
	errResponse := &Error{Message: err.Error()}
	result, err := json.Marshal(errResponse)
	if err != nil {
		panic(err)
	}
	return string(result)
}

// GenerateMAC is pulled from https://stackoverflow.com/questions/21018729/generate-mac-address-in-go
func GenerateMAC() (net.HardwareAddr, error) {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		fmt.Println("error:", err)
		return nil, err
	}

	buf[0] &= 254
	buf[0] |= 2

	return buf, nil
}

// HashPassword encrypts a plaintext string and returns the hashed version
func HashPassword(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hashed)
}

// MustGetInt will get an integer from ninjarouter URL or panic
func MustGetInt(r *http.Request, name string) int {
	nameStr := ninjarouter.Var(r, name)
	nameInt, err := strconv.Atoi(nameStr)
	if err != nil {
		panic(err)
	}
	return nameInt
}

// MustGetString will get an string from ninjarouter URL or panic
func MustGetString(r *http.Request, name string) string {
	nameStr := ninjarouter.Var(r, name)
	return nameStr
}

// MustDecodeJSON receives a pointer to struct, and updates the struct values
// with the values from the JSON in the http request
func MustDecodeJSON(body io.Reader, target interface{}) {
	err := json.NewDecoder(body).Decode(target)
	if err != nil {
		panic(errors.Wrap(err, "Could not decode JSON"))
	}
}

// MustMarshalJSON will return a byte array representation of a struct
func MustMarshalJSON(input interface{}) []byte {
	response, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}

	return response
}

// RaijinTemplate is the values needed to create a container based on a template
type RaijinTemplate struct {
	OS         string `json:"os"`
	OSVersion  string `json:"osVersion"`
	Name       string `json:"Name"`
	OSVersion2 string `json:"osVersion2"`
	Arch       string `json:"arch"`
	Extension  string `json:"extension"`
}

// ParseTemplate will convert a proxmox template string into a struct
func ParseTemplate(templateStr string) (*RaijinTemplate, error) {
	expr := `\/(.*?){1}-(.*?){1}-(.*?){1}_(.*?){1}_(.*?){1}\.`

	r := regexp.MustCompile(expr)
	result := r.FindStringSubmatch(templateStr)
	template := &RaijinTemplate{
		OS:         result[1],
		OSVersion:  result[2],
		Name:       result[3],
		OSVersion2: result[4],
		Arch:       result[5],
	}

	if result[0] == "" ||
		result[1] == "" ||
		result[2] == "" ||
		result[3] == "" ||
		result[4] == "" ||
		result[5] == "" {
		return nil, errors.New("no matches found from template")
	}
	return template, nil
}
