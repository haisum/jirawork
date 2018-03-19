package activity

import (
	"encoding/xml"
	"fmt"
	"github.com/haisum/jirawork/pkg/jirahttp"
	"io/ioutil"
	"net/http"
)

const (
	url        = "%s/activity?streams=user+IS+%s&os_authType=basic&maxResults=%d&activity+IS+issue:update&streams=update-date+BETWEEN+%d+%d"
	maxResults = 100
)

type Client interface {
	Get(after, before int64, v interface{}) error
}

type client struct {
	Config jirahttp.Config
}

func NewClient(config jirahttp.Config) *client {
	return &client{config}
}

func (c *client) Get(after, before int64, v interface{}) error {
	actionURL := fmt.Sprintf(url, c.Config.URL, c.Config.Username, maxResults, after, before)
	c.Config.Log("url", actionURL)
	req,_ := http.NewRequest("GET", actionURL, nil)
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		b, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("http error code: %d, body: %s", resp.StatusCode, b)
	}
	//b, _ := ioutil.ReadAll(resp.Body)
	decoder := xml.NewDecoder(resp.Body)
	err = decoder.Decode(v)
	return err
}
