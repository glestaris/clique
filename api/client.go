package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type Client struct {
	host string
	port uint16

	httpClient *http.Client
}

func NewClient(host string, port uint16, timeout time.Duration) *Client {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: timeout,
		}).Dial,
		DisableKeepAlives: true,
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	return &Client{
		host: host,
		port: port,

		httpClient: httpClient,
	}
}

func (c *Client) Ping() error {
	if _, err := c.do("get", "ping", nil); err != nil {
		return err
	}

	return nil
}

func (c *Client) TransfersByState(state TransferState) ([]Transfer, error) {
	data, err := c.do("get", fmt.Sprintf("transfers/%s", state.String()), nil)
	if err != nil {
		return nil, err
	}

	var res []Transfer
	if err := json.Unmarshal(data, &res); err != nil {
		// untested return
		return nil, fmt.Errorf("unmarshalling JSON: %s", err)
	}

	return res, nil
}

func (c *Client) TransferResults() ([]TransferResults, error) {
	data, err := c.do("get", "transfer_results", nil)
	if err != nil {
		return nil, err
	}

	var res []TransferResults
	if err := json.Unmarshal(data, &res); err != nil {
		// untested return
		return nil, fmt.Errorf("unmarshalling JSON: %s", err)
	}

	return res, nil
}

func (c *Client) TransferResultsByIP(ip net.IP) ([]TransferResults, error) {
	data, err := c.do("get", fmt.Sprintf("transfer_results/%s", ip), nil)
	if err != nil {
		return nil, err
	}

	var res []TransferResults
	if err := json.Unmarshal(data, &res); err != nil {
		// untested return
		return nil, fmt.Errorf("unmarshalling JSON: %s", err)
	}

	return res, nil
}

func (c *Client) CreateTransfer(spec TransferSpec) error {
	if _, err := c.do("post", "transfers", spec); err != nil {
		return err
	}

	return nil
}

func (c *Client) route(path string) string {
	return fmt.Sprintf("http://%s:%d/%s", c.host, c.port, path)
}

func (c *Client) do(method, path string, req interface{}) ([]byte, error) {
	var (
		resp *http.Response
		err  error
	)
	if method == "get" {
		resp, err = c.httpClient.Get(c.route(path))
	} else if method == "post" {
		data, err := json.Marshal(req)
		if err != nil {
			// untested return
			return nil, fmt.Errorf("invalid request: %s", err)
		}

		resp, err = c.httpClient.Post(
			c.route(path),
			"application/json",
			bytes.NewBuffer(data),
		)
	} else {
		// untested return
		return nil, fmt.Errorf("unknown method '%s'", method)
	}

	if err != nil {
		return nil, fmt.Errorf("making request: %s", err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// untested return
		return nil, fmt.Errorf("reading error response: %s", err)
	}

	if resp.StatusCode == 200 {
		return data, nil
	}

	var res ServerError
	if err := json.Unmarshal(data, &res); err != nil {
		// untested return
		return nil, fmt.Errorf("reading error response: %s", err)
	}

	return nil, fmt.Errorf(res.Msg)
}
