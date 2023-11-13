package pgo

import (
	"fmt"
	"github.com/viant/pgo/build"
	"os"
	"path"
	"strings"
)

// Options options
type Options struct {
	SourceURL    []string          `short:"s" long:"src" description:"plugin source project location"  `
	DestURL      string            `short:"d" long:"dest" description:"plugin dest location"  `
	Name         string            `short:"n" long:"name" description:"plugin name, default main"  `
	Arch         string            `short:"a" long:"arch" description:"amd64|arm64"  `
	Os           string            `short:"o" long:"os" description:"linux|darwin"  `
	Version      string            `short:"v" long:"ver" description:"go version"  `
	ModPath      string            `short:"m" long:"modpath" description:"go mod path"  `
	MainPath     string            `short:"p" long:"mainpath" description:"main path in project"  `
	BuildArgs    []string          `short:"b" long:"barg" description:"build args" `
	LdFlags      string            `short:"f" long:"f" description:"ldflags" `
	Env          map[string]string `short:"e" long:"env" description:"env variables" `
	BuildMode    string            `short:"b" long:"bmode" description:"build mode" choice:"exec" choice:"plugin" `
	UseContainer bool              `short:"N" long:"container" description:"use container to ensure consistency" `
	Compression  string            `short:"c" long:"compression" description:"compression codec" choice:"gzip"  `
	WithLogger   bool              `short:"l" long:"log" description:"with debug logger" `
	GoPath       string            `short:"p" long:"gopath" description:"go path" `
	GoRoot       string            `short:"r" long:"goroot" description:"go root" `
}

// Validate check if option are valid
func (o *Options) Validate() error {
	if len(o.SourceURL) == 0 {
		return fmt.Errorf("sourceURL was empty")
	}
	if o.DestURL == "" {
		return fmt.Errorf("DestURL was empty")
	}
	if o.Version == "" {
		return fmt.Errorf("version was empty")
	}
	if o.Arch == "" {
		return fmt.Errorf("arch was empty")
	}
	if o.Os == "" {
		return fmt.Errorf("os was empty")
	}

	return nil
}

// buildSpec return build.Build specification
func (o *Options) buildSpec() *build.Build {
	ret := &build.Build{}
	ret.Go.Runtime.Os = o.Os
	ret.Go.Runtime.Arch = o.Arch
	ret.Go.Runtime.Version = o.Version
	ret.Go.Env = o.Env
	ret.Go.Root = o.GoRoot
	ret.Go.Path = o.GoPath

	if len(o.SourceURL) > 0 {
		ret.Source.URL = o.SourceURL[0]
		for i := 1; i < len(o.SourceURL); i++ {
			ret.LocalDep = append(ret.LocalDep, &build.Source{URL: o.SourceURL[i]})
		}
	}
	ret.Name = o.Name
	ret.Spec.ModPath = o.ModPath
	ret.Spec.ModPath = o.ModPath
	ret.Spec.BuildArgs = o.BuildArgs
	ret.Go.LdFlags = o.LdFlags
	ret.Spec.MainPath = o.MainPath
	ret.Mode = o.BuildMode
	ret.Go.Runtime.UseContainer = o.UseContainer
	return ret
}

// Init initialises option
func (o *Options) Init() {
	for i := range o.SourceURL {
		o.SourceURL[i] = normalizeLocation(o.SourceURL[i])
	}
	o.DestURL = normalizeLocation(o.DestURL)
}

func normalizeLocation(location string) string {
	if strings.HasPrefix(location, "~") {
		return os.Getenv("HOME") + location[1:]
	}
	if !strings.Contains(location, ":/") && !strings.HasPrefix(location, "/") {
		cwd, _ := os.Getwd()
		return path.Join(cwd, location)
	}
	return location
}
