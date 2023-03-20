package pgo

import (
	"context"
	"github.com/viant/afs"
	"github.com/viant/pgo/build"
	"github.com/viant/pgo/internal/builder"
)

//Build builds plugin
func Build(option *Options, opts ...build.Option) error {
	option.Init()
	if err := option.Validate(); err != nil {
		return err
	}
	var aBuilder = builder.New(&builder.Config{}, builder.WithLinuxAmd64)
	spec := option.buildSpec()
	module, err := aBuilder.Build(context.Background(), spec, opts...)
	if err != nil {
		return err
	}
	module.Compression = option.Compression
	fs := afs.New()
	return module.Store(context.Background(), fs, option.DestURL)
}
