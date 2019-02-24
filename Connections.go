package connect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// Connections will list the connections of displayName. If displayName is
// empty, the current authenticated users connection list wil be returned.
func (c *Client) Connections(displayName string) ([]SocialProfile, error) {
	// There also exist an endpoint without /pagination/ but it will return
	// 403 for *some* connections.
	URL := "https://connect.garmin.com/modern/proxy/userprofile-service/socialProfile/connections/pagination/" + displayName

	if !c.authenticated() && displayName == "" {
		return nil, ErrNotAuthenticated
	}

	var proxy struct {
		Connections []SocialProfile `json:"userConnections"`
	}

	err := c.getJSON(URL, &proxy)
	if err != nil {
		return nil, err
	}

	return proxy.Connections, nil
}

// PendingConnections returns a list of pending connections.
func (c *Client) PendingConnections() ([]SocialProfile, error) {
	URL := "https://connect.garmin.com/modern/proxy/userprofile-service/connection/pending"

	if !c.authenticated() {
		return nil, ErrNotAuthenticated
	}

	pending := make([]SocialProfile, 0, 10)

	err := c.getJSON(URL, &pending)
	if err != nil {
		return nil, err
	}

	return pending, nil
}

// AcceptConnection will accept a pending connection.
func (c *Client) AcceptConnection(connectionRequestID int) error {
	URL := fmt.Sprintf("https://connect.garmin.com/modern/proxy/userprofile-service/connection/accept/%d", connectionRequestID)
	payload := struct {
		ConnectionRequestID int `json:"connectionRequestId"`
	}{
		ConnectionRequestID: connectionRequestID,
	}

	body := bytes.NewBuffer(nil)
	enc := json.NewEncoder(body)
	err := enc.Encode(payload)
	if err != nil {
		return err
	}

	req, err := c.newRequest("PUT", URL, body)
	if err != nil {
		return err
	}

	req.Header.Add("nk", "NT")
	req.Header.Add("content-type", "application/json")

	_, err = c.do(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) SearchConnections(keyword string) ([]SocialProfile, error) {
	URL := "https://connect.garmin.com/modern/proxy/usersearch-service/search"

	payload := url.Values{
		"start":   {"1"},
		"limit":   {"20"},
		"keyword": {keyword},
	}

	req, err := c.newRequest("POST", URL, strings.NewReader(payload.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("nk", "NT")
	req.Header.Add("content-type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var proxy struct {
		Profiles []SocialProfile `json:"profileList"`
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&proxy)
	if err != nil {
		return nil, err
	}

	return proxy.Profiles, nil
}

// RemoveConnection will remove a connection.
func (c *Client) RemoveConnection(connectionRequestID int) error {
	URL := fmt.Sprintf("https://connect.garmin.com/modern/proxy/userprofile-service/connection/end/%d", connectionRequestID)

	req, err := c.newRequest("PUT", URL, nil)
	if err != nil {
		return err
	}

	req.Header.Add("nk", "NT")

	_, err = c.do(req)

	return err
}
