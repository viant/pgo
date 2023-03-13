# pgo (Go plugin)

[![GoReportCard](https://goreportcard.com/badge/github.com/viant/pgo)](https://goreportcard.com/report/github.com/viant/pgo)
[![GoDoc](https://godoc.org/github.com/viant/pgo?status.svg)](https://godoc.org/github.com/viant/pgo)

This library is compatible with Go 1.17+

Please refer to [`CHANGELOG.md`](CHANGELOG.md) if you encounter breaking changes.

- [Motivation](#motivation)
- [Usage](#usage)
- [Contribution](#contributing-to-pgo)
- [License](#license)

## Motivation

The goal of this library is to simplify build and management GoLang plugins.

Go plugin comes with the following constrains,
- go version, os and cpu architecture match
- go.mod dependencies have to be the same
- once plugins is loaded can not be reloaded or unloaded


Since GoLang plugins rely on shared libraries in order to address the os,cpu architecture match, 
with this project you can build ahead of time the same plugin for various platform and go version, 
the build plugin uses the following naming:
NAME_GO_VERSION_GOOS_GOARCH.so

This project allow you to rebuild the same plugin source and reload it as long there is available memory,
note that previous loaded plugins stay in memory.

Each plugin is stamped by SCN time based sequence change number which allows to load the newest version of the plugin. 

Time based sequence change number allows application build/design with "initial SCN version", that 
can be supplemented with plugin above that version.


## Usage

#### Building plugin
To build plugin:
```go
package pgo_test

import (
	"github.com/viant/pgo"
	"log"
)
//Example_Build build plugin example
func Example_Build() {
	options := &pgo.Options{
		SourceURL:  "internal/builder/testdata/my.zip",
		DestURL:    "~/plugin/",
		Os:         "linux",
		Arch:       "amd64",
		Version:    "1.17.6",
		Name:       "main",
		WithLogger: true,
	}
	err := pgo.Build(options)
	if err != nil {
		log.Fatalln(err)
	}
}
```

### Load plugins
```go
package example

import (
	"github.com/viant/pgo/build"
	"github.com/viant/pgo/manager"
	"path"
	"time"
	"context"
    "fmt"
)
func Example_Open() {
	
	var buildTime time.Time
	var initScn = build.AsScn(buildTime)
	srv := manager.New(initScn)
	runtime := build.NewRuntime()
	pluginName := runtime.PluginName("main.so")
	basePluginPath := "/var/some_path"
	goPlugin, err := srv.Open(context.Background(), path.Join(basePluginPath, pluginName))
	if manager.IsPluginToOld(err) {
		//plugin not needed, already the newest version is loaded or initial SCN is more recent
	}
	if goPlugin != nil {
		xx, err := goPlugin.Lookup("someSymbol")
		fmt.Printf("%T(%v) %v\n", xx, xx, err)
	}
}


```


## Contributing to pgo

pgo is an open source project and contributors are welcome!

See [TODO](TODO.md) list

## Credits and Acknowledgements

**Library Author:** Adrian Witas, Kamil Larysz

