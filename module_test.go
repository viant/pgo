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
		WithLogger: true,
	}
	err := pgo.Build(options)
	if err != nil {
		log.Fatalln(err)
	}
}
