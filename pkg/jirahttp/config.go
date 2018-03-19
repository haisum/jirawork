package jirahttp

import (
	"errors"
)

type Config struct {
	URL      string
	Username string
	Password string
	Log      func(vals ...interface{})
}

func NewConfig(url, username, password string, log func(vals ...interface{})) (Config, error) {
	config := Config{
		url,
		username,
		password,
		log,
	}
	if username == "" || password == "" || url == "" {
		return config, errors.New("jira username, password and url are required")
	}
	if config.Log == nil {
		config.Log = func(vals ...interface{}) {}
	}
	return config, nil
}
