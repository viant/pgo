package builder_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/viant/pgo/build"
	"github.com/viant/pgo/internal/builder"
	"github.com/viant/pgo/internal/endpoint"
	"path/filepath"
	"testing"
	"time"
)

func TestClient_Build(t *testing.T) {

	aBuilder := builder.New(&builder.Config{})
	router := endpoint.NewRouter(aBuilder)
	server := endpoint.NewServer(8082, router)
	ctx := context.Background()

	go server.Start()
	defer server.Shutdown(ctx)
	time.Sleep(time.Second)
	aClient := builder.NewClient("http://127.0.0.1:8082")

	pluginBaseURL := filepath.Join("../builder/testdata")
	pluginURL := filepath.Join(pluginBaseURL, "my.zip")

	runtime := build.NewRuntime()
	resp, err := aClient.Build(ctx, build.New(pluginURL, runtime))
	assert.Nil(t, err)
	assert.True(t, len(resp.Data) > 0)
}
