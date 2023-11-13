package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/viant/afs"
	"github.com/viant/afs/url"
	"github.com/viant/pgo/build"
	"github.com/viant/pgo/internal"
	"io"
	"net/http"
	"strings"
)

// Client represetns a client
type Client struct {
	fs      afs.Service
	BaseURL string
}

// IsUp returns true if client is up
func (c *Client) IsUp() bool {
	URL := url.Join(c.BaseURL, internal.StatusURI)
	response, err := http.DefaultClient.Get(URL)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	data, _ := io.ReadAll(response.Body)
	return string(data) == "ok"
}

// Build builds plugin
func (c *Client) Build(ctx context.Context, buildSpec *build.Build) (*build.Module, error) {
	buildSpec.Init()
	if err := buildSpec.Validate(); err != nil {
		return nil, err
	}
	if err := buildSpec.Source.Pack(ctx, c.fs); err != nil {
		return nil, err
	}
	reqData, err := json.Marshal(buildSpec)
	if err != nil {
		return nil, err
	}
	URL := url.Join(c.BaseURL, internal.BuildURI)
	response, err := http.DefaultClient.Post(URL, "application/json", bytes.NewReader(reqData))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	ret := &build.Module{}
	if !json.Valid(data) {
		return ret, fmt.Errorf("%s", data)
	}
	err = json.Unmarshal(data, ret)
	return ret, err
}

// NewClient creates a client
func NewClient(baseURL string) *Client {
	baseURL = strings.Trim(baseURL, "/")
	return &Client{BaseURL: baseURL, fs: afs.New()}
}
