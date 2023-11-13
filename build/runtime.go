package build

import (
	"fmt"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

// Runtime represents a go runtime
type Runtime struct {
	Arch         string
	Os           string
	Version      string
	UseContainer bool
}

// Init initialises plugin
func (r *Runtime) Init() {
	if r.Os == "" {
		r.Os = runtime.GOOS
	}
	if r.Arch == "" {

		cmd := exec.Command("uname", "-m")
		if data, err := cmd.Output(); err == nil {
			if strings.Contains(string(data), "x86") {
				r.Arch = "amd64"
			}
		}
		if r.Arch == "" {
			r.Arch = runtime.GOARCH
		}

	}
	if r.Version == "" {
		r.Version = strings.Replace(runtime.Version(), "go", "", 1)
	}
}

// DetectVersion detect runtime version or fallback to build go version in case of error
func (r *Runtime) DetectVersion() {
	r.Init()
	command := exec.Command("go", "version")
	if output, err := command.CombinedOutput(); err == nil {
		fragment := strings.TrimSpace(string(output))
		if index := strings.LastIndex(fragment, "go"); index != -1 {
			fragment = fragment[index+2:]
			if index := strings.Index(fragment, " "); index != -1 {
				fragment = fragment[:index]
			}
			r.Version = fragment
		}
	}
}

// ValidateOsAndArch checks if runtime os and arch is compatible
func (r *Runtime) ValidateOsAndArch(runtime *Runtime) error {
	if r.Arch != runtime.Arch {
		return fmt.Errorf("invalid plugin arch: expected: %v, but had: %v", r.Arch, runtime.Arch)
	}
	if r.Os != runtime.Os {
		return fmt.Errorf("invalid plugin os: expected: %v, but had: %v", r.Os, runtime.Os)
	}
	return nil
}

// Validate checks if runtime is compatible
func (r *Runtime) Validate(runtime *Runtime) error {
	if err := r.ValidateOsAndArch(runtime); err != nil {
		return err
	}
	if r.Version != runtime.Version {
		return fmt.Errorf("invalid plugin go version: expected: %v, but had: %v", r.Version, runtime.Version)
	}
	return nil
}

// PluginName returns runtime specific plugin name
func (r *Runtime) PluginName(name string) string {
	if name == "" {
		name = "main.so"
	}
	compressed := strings.HasSuffix(name, ".gz")
	if compressed {
		name = name[:len(name)-3]
	}

	var adjusted = name
	if ext := path.Ext(adjusted); ext != "" {
		adjusted = adjusted[:len(adjusted)-len(ext)]
	}
	ret := adjusted + "_" + strings.ReplaceAll(r.Version, ".", "_") + "_" + r.Os + "_" + r.Arch + ".so"
	if compressed {
		ret += ".gz"
	}
	return ret
}

// PluginName returns runtime specific plugin name
func (r *Runtime) InfoName(name string) string {
	if name == "" {
		name = "main.pinf"
	}
	var adjusted = name
	if ext := path.Ext(adjusted); ext != "" {
		adjusted = adjusted[:len(adjusted)-len(ext)]
	}
	return adjusted + "_" + strings.ReplaceAll(r.Version, ".", "_") + "_" + r.Os + "_" + r.Arch + ".pinf"
}

// NewRuntime creates a runtime
func NewRuntime() Runtime {
	ret := Runtime{}
	ret.Init()
	return ret
}
