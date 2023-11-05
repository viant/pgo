package build

import (
	"context"
	"fmt"
	"github.com/viant/afs"
	"log"
	"runtime"
)

type (
	//Build represents build spec
	Build struct {
		Name     string
		Source   Source
		LocalDep []*Source
		Spec     Spec
		Mode     string
		Go       GoBuild
		logger   func(format string, args ...interface{})
	}

	//Option represents build option
	Option func(b *Build)

	//GoBuild represents go build spec
	GoBuild struct {
		Runtime
		LdFlags string
		Env     map[string]string
		Path    string
		Root    string
	}

	//Spec represents plygin spec
	Spec struct {
		ModPath   string
		MainPath  string
		BuildArgs []string
	}
)

// WithLogger create logger build option
func WithLogger(printer func(format string, args ...interface{})) func(b *Build) {
	if printer == nil {
		printer = log.Printf
	}
	return func(b *Build) {
		b.logger = func(format string, args ...interface{}) {
			printer(format, args...)
		}
	}
}

// Logf logs
func (b *Build) Logf(format string, args ...interface{}) {
	if b.logger == nil {
		return
	}
	b.logger(format, args...)
}

// Init checks if build is valid
func (b *Build) Init() {
	b.Go.Init()
	if b.Mode == "" {
		b.Mode = "plugin"
	}
}

// Validate check if build is valid
func (b *Build) Validate() error {
	return b.Go.Validate()
}

// Init  initialises build
func (b *GoBuild) Init() {
	if b == nil {
		return
	}
	b.Runtime.Init()
	if b.Os == "" {
		b.Os = runtime.GOOS
	}
	if b.Arch == "" {
		b.Os = runtime.GOARCH
	}

}

// Validate validates if go build is valid
func (b *GoBuild) Validate() error {
	if b.Version == "" {
		return fmt.Errorf("go.Version was empty")
	}
	fs := afs.New()
	if b.Root != "" {
		if ok, _ := fs.Exists(context.Background(), b.Root); !ok {
			return fmt.Errorf("GOROOT does not exists")
		}
	}
	return nil
}

// New creates build options
func New(URL string, runtime Runtime, opts ...Option) *Build {
	ret := &Build{}
	ret.Source.URL = URL
	ret.Go.Runtime = runtime
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}
