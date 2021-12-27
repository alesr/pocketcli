package pocketcli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	endpointRetrieve        = "/get"
	xErrorHeader     string = "X-Error"
)

type (
	httpClient interface {
		Do(*http.Request) (*http.Response, error)
	}

	Client struct {
		host         string
		consumerKey  string
		accessToken  string
		username     string
		httpCli      httpClient
		retrieveOpts *bytes.Buffer
	}

	RetrieveResult struct {
		List map[string]Bookmark `json:"list"`
	}

	Bookmark struct {
		ID    int    `json:"item_id,string"`
		Title string `json:"resolved_title"`
		URL   string `json:"resolved_url"`
	}
)

func New(httpCli httpClient, host, consumerKey, accessToken, username string) (*Client, error) {
	retrieveOpts := struct {
		Tag string `json:"tag,omitempty"`
	}{
		Tag: "rmk",
	}

	opts, err := json.Marshal(retrieveOpts)
	if err != nil {
		return nil, fmt.Errorf("could not marshal retrieve options: %w", err)
	}

	return &Client{
		host:         host,
		consumerKey:  consumerKey,
		accessToken:  accessToken,
		username:     username,
		httpCli:      httpCli,
		retrieveOpts: bytes.NewBuffer(opts),
	}, nil
}

func (c *Client) Retrieve(ctx context.Context) (map[string]Bookmark, error) {
	req, err := http.NewRequest(http.MethodPost, c.host+endpointRetrieve, c.retrieveOpts)
	if err != nil {
		return nil, fmt.Errorf("could not create request for fetch bookmarks: %w", err)
	}

	resp, err := c.do(ctx, req, c.accessToken, http.StatusOK)
	if err != nil {
		return nil, fmt.Errorf("could not fetch bookmarks: %w", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}
	defer resp.Body.Close()

	var result RetrieveResult

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response body: %w", err)
	}
	return result.List, nil
}

func (c *Client) do(ctx context.Context, req *http.Request, accessTkn string, expectedStatusCode int) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/json charset=utf-8")

	q := req.URL.Query()
	q.Add("consumer_key", c.consumerKey)
	q.Add("access_token", accessTkn)
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not make request: %w", err)
	}

	if resp.StatusCode != expectedStatusCode {
		return nil, fmt.Errorf(
			"unexpected status code '%s': %s",
			http.StatusText(resp.StatusCode),
			resp.Header.Get(xErrorHeader),
		)
	}
	return resp, nil
}
