package pgo_test

import (
	"github.com/viant/pgo"
	"github.com/viant/pgo/build"
	"log"
	"testing"
)

//Example_Build build plugin example
func Example_Build() {
	options := &pgo.Options{
		SourceURL:  []string{"internal/builder/testdata/my.zip"},
		DestURL:    "~/plugin/",
		Os:         "linux",
		Arch:       "amd64",
		Version:    "1.17.6",
		WithLogger: true,
	}
	err := pgo.Build(options)
	if err != nil {
		log.Fatalln(err)
	}
}

//Example_Build build plugin example
func Test_Info(t *testing.T) {
	options := &pgo.Options{
		SourceURL: []string{
			"/Users/awitas/go/src/github.vianttech.com/adelphic/datlydev/pkg",
		},
		Name:        "main.so",
		DestURL:     "~/ppp/",
		Compression: "gzip",
		Os:          "linux",
		Arch:        "amd64",
		Version:     "1.17.6",
		BuildMode:   "plugin",
		BuildArgs:   []string{"-trimpath"},
		LdFlags:     "-X main.BuildTimeInS=1678941563",
		WithLogger:  true,
	}
	err := pgo.Build(options, build.WithLogger(nil))
	if err != nil {
		log.Fatalln(err)
	}
}
