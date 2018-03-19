package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/haisum/jirawork/pkg/jirahttp"
	"io/ioutil"
	"net/http"
)

const (
	url = "%s%s"
)

type Client interface {
	Post(endpoint string, v interface{}) error
}

type client struct {
	Config jirahttp.Config
}

func NewClient(config jirahttp.Config) *client {
	return &client{config}
}

func (c *client) Post(endpoint string, v interface{}) error {
	actionURL := fmt.Sprintf(url, c.Config.URL, endpoint)
	c.Config.Log("url", actionURL)
	body, err := json.Marshal(v)
	c.Config.Log("params", string(body[:]))
	if err != nil {
		return err
	}
	r := bytes.NewReader(body)
	req,_ := http.NewRequest("POST", actionURL, r)
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("http error code: %d, body: %s", resp.StatusCode, b)
	}
	return nil
}
