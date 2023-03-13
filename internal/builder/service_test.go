package builder

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/viant/afs"
	"github.com/viant/pgo/build"
	"path/filepath"
	"testing"
)

func TestService(t *testing.T) {
	basePath := "testdata"
	source := build.Source{URL: filepath.Join(basePath, "my.zip")}
	fs := afs.New()
	ctx := context.Background()
	err := source.Pack(ctx, fs)
	assert.Nil(t, err)
	srv := New(&Config{})
	aPlugin, err := srv.Build(ctx, &build.Build{
		Go:     build.GoBuild{Runtime: build.Runtime{Version: "1.17.6"}},
		Source: source,
	})
	if !assert.Nil(t, err) {
		fmt.Println(err.Error())
		return
	}
	err = aPlugin.Store(ctx, fs, filepath.Join(basePath, "dist"))
	if !assert.Nil(t, err) {
		fmt.Println(err.Error())
	}
}
