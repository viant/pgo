package main

import (
	flags "github.com/jessevdk/go-flags"
	"github.com/viant/pgo"
	"github.com/viant/pgo/build"

	"log"
	"os"
)

func main() {
	args := os.Args[1:]
	options := &pgo.Options{}
	if _, err := flags.ParseArgs(options, args); err != nil {
		log.Fatalln(err)
	}
	var opts = []build.Option{}
	if options.WithLogger {
		opts = append(opts, build.WithLogger(nil))
	}
	if err := pgo.Build(options, opts...); err != nil {
		log.Fatalln(err)
	}
}
