package pgo_test

import (
	"github.com/viant/pgo"
	"log"
)

//Example_Build build plugin example
func Example_Build() {
	options := &pgo.Options{
		SourceURL: []string{
			"/Users/awitas/go/src/github.vianttech.com/adelphic/datlydev/.build/datly",
			"/Users/awitas/go/src/github.vianttech.com/adelphic/datlydev/pkg",
		},
		DestURL:    "~/dddd/",
		Os:         "darwin",
		Arch:       "amd64",
		Version:    "1.17.6",
		BuildMode:  "exec",
		WithLogger: true,
	}
	err := pgo.Build(options)
	if err != nil {
		log.Fatalln(err)
	}
}
