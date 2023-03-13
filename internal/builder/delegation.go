package builder

import (
	"context"
	"fmt"
	"github.com/viant/afs"
	"github.com/viant/pgo/build"
	"path"
)

//Delegation represents a delegation
type Delegation struct {
	build.Runtime
	Name  string
	Image string
	Port  int
}

var searchPath = []string{"/usr/local/bin", "/usr/bin", "/bin", "/usr/sbin", "/sbin"}

func (d Delegation) findDocker(customPath string) string {
	fs := afs.New()
	if customPath != "" {
		location := path.Join(customPath, "docker")
		if ok, _ := fs.Exists(context.Background(), location); ok {
			return location
		}
	}
	for _, candidate := range searchPath {
		location := path.Join(candidate, "docker")
		if ok, _ := fs.Exists(context.Background(), location); ok {
			return location
		}
	}
	return ""
}

func (d *Delegation) baseURL() string {
	return fmt.Sprintf("http://127.0.0.1:%v", d.Port)
}

func (d *Delegation) buildDockerRunCmdArgs(dockerCustomPath string) (string, []string, error) {
	cmd := d.findDocker(dockerCustomPath)
	if cmd == "" {
		return "", nil, fmt.Errorf("failed to locate docker, please make sure it's installed")
	}
	args := append([]string{}, "run",
		"-d",
		"--name", d.Name,
		"--platform", fmt.Sprintf("%v/%v", d.Os, d.Arch),
		"-p", fmt.Sprintf("%v:%v", d.Port, d.Port),
		d.Image)
	return cmd, args, nil
}

func (d *Delegation) buildDockerStartCmdArgs(dockerCustomPath string) (string, []string, error) {
	cmd := d.findDocker(dockerCustomPath)
	if cmd == "" {
		return "", nil, fmt.Errorf("failed to locate docker, please make sure it's installed")
	}
	args := append([]string{}, "start", d.Name)
	return cmd, args, nil
}

//Match matches delegation with supplied runtime
func (d Delegations) Match(runtime *build.Runtime) *Delegation {
	if d == nil || len(d) == 0 {
		return nil
	}
	for _, candidate := range d {
		if candidate.ValidateOsAndArch(runtime) == nil {
			return candidate
		}
	}
	return nil
}
