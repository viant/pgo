package builder

import (
	"github.com/viant/pgo/build"
)

type (
	//Config represents config
	Config struct {
		Runtime     build.Runtime
		delegations Delegations
		dockerPath  string
	}

	//Delegations represents delegation
	Delegations []*Delegation

	//Option represents config option
	Option func(c *Config)
)

//WithLinuxAmd64 create linux amd docker delegation
func WithLinuxAmd64(c *Config) {
	delegated := &Delegation{Runtime: build.Runtime{Os: "linux", Arch: "amd64"}, Image: "viant/pgo:latest", Name: "pgo", Port: 8089}
	c.delegations = append(c.delegations, delegated)
}

//WithDockerPath creates docker path
func WithDockerPath(path string) func(c *Config) {
	return func(c *Config) {
		c.dockerPath = path
	}
}

//NewConfig creates a config
func NewConfig() *Config {
	cfg := &Config{}
	cfg.Runtime.Init()
	return cfg
}
