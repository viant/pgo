package manager_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs"
	"github.com/viant/pgo/build"
	"github.com/viant/pgo/internal/builder"
	"github.com/viant/pgo/manager"
	"path"
	"path/filepath"
	"testing"
	"time"
)

func TestServiceLoad(t *testing.T) {

	serviceTestBase := filepath.Join("testdata")
	pluginTestBase := filepath.Join("../internal/builder", "testdata")
	var testCases = []struct {
		description string
		sourceURL   string
		symbolName  string
	}{
		{
			description: "plugin load test",
			sourceURL:   filepath.Join(pluginTestBase, "my.zip"),
			symbolName:  "Contract",
		},
	}

	cfg := builder.NewConfig()
	aBuilder := builder.New(cfg)
	ctx := context.Background()
	fs := afs.New()
	initScn := build.AsScn(time.Now())
	for _, testCase := range testCases {
		aPlugin, err := aBuilder.Build(ctx, build.New(testCase.sourceURL, cfg.Runtime), build.WithLogger(nil))
		if !assert.Nil(t, err, testCase.description) {
			continue
		}
		destPlugin := filepath.Join(serviceTestBase, "dist")
		aPlugin.Compression = "gzip"
		err = aPlugin.Store(ctx, fs, destPlugin)
		if !assert.Nil(t, err, testCase.description) {
			continue
		}

		srv := manager.New(initScn)
		assert.NotNil(t, srv, testCase.description)
		infoName := cfg.Runtime.InfoName("main.pinf")
		goPlugin, err := srv.OpenWithInfoURL(ctx, path.Join(destPlugin, infoName))
		if !assert.Nil(t, err, testCase.description) {
			continue
		}
		pluginName := cfg.Runtime.PluginName("main.so.gz")
		//second time would skip loading the plugin
		_, err = srv.Open(ctx, path.Join(destPlugin, pluginName))
		assert.True(t, manager.IsPluginToOld(err))
		value, err := goPlugin.Lookup(testCase.symbolName)
		if !assert.Nil(t, err, testCase.description) {
			continue
		}
		if !assert.NotNil(t, value, testCase.description) {
			continue
		}

		//build again plugin and load it againg
		aPlugin, err = aBuilder.Build(ctx, build.New(testCase.sourceURL, cfg.Runtime))
		if !assert.Nil(t, err, testCase.description) {
			continue
		}
		aPlugin.Compression = "gzip"
		err = aPlugin.Store(ctx, fs, destPlugin)
		if !assert.Nil(t, err, testCase.description) {
			continue
		}
		goPlugin, err = srv.Open(ctx, path.Join(destPlugin, pluginName))
		if !assert.Nil(t, err, testCase.description) {
			continue
		}
		assert.NotNil(t, goPlugin, testCase.description)

	}

}

func ExampleLoad() {
	var buildTime time.Time
	var initScn = build.AsScn(buildTime)
	srv := manager.New(initScn)
	runtime := build.NewRuntime()
	pluginName := runtime.PluginName("main.so")
	basePluginPath := "/tmp/plugin"
	goPlugin, err := srv.Open(context.Background(), path.Join(basePluginPath, pluginName))
	if manager.IsPluginToOld(err) {
		//plugin not needed, already newst version is loaded or initScn is more recent
	}
	if goPlugin != nil {
		xx, err := goPlugin.Lookup("someSymbol")
		fmt.Printf("%T(%v) %v\n", xx, xx, err)
	}

}
