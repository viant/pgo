package builder

import (
	"fmt"
	"github.com/viant/pgo/build"
	"os"
	"os/exec"
	"strings"
)

func (s *Service) runDocker(delegation *Delegation, buildSpec *build.Build) error {
	cmd, args, err := delegation.buildDockerStartCmdArgs(s.cfg.dockerPath)
	if err != nil {
		return err
	}
	stdout, err := s.runCommand(cmd, args, buildSpec)
	if err == nil && !strings.Contains(strings.ToLower(stdout), "error") {
		return nil
	}
	cmd, args, err = delegation.buildDockerRunCmdArgs(s.cfg.dockerPath)
	if err != nil {
		return err
	}
	stdout, err = s.runCommand(cmd, args, buildSpec)
	if err != nil {
		return err
	}
	if strings.Contains(strings.ToLower(stdout), "error") {
		return fmt.Errorf(stdout)
	}
	return nil
}

func (s *Service) runCommand(cmd string, args []string, buildSpec *build.Build) (string, error) {
	command := exec.Command(cmd, args...)
	command.Env = os.Environ()
	buildSpec.Logf("%v", command.String())
	output, err := command.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
