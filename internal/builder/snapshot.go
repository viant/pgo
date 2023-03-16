package builder

import (
	"bytes"
	"fmt"
	"github.com/viant/pgo/build"
	"golang.org/x/mod/modfile"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const defaultDirPermission = 0o777

//Snapshot represent a plugin snapshot
type Snapshot struct {
	Created   time.Time
	buildMode string
	Spec      *build.Spec
	build.GoBuild
	PluginDestPath  string
	PluginBuildPath string
	ModFile         *modfile.File
	BuildModPath    string
	ModFiles        []*modfile.File
	Mains           []string
	BaseDir         string
	TempDir         string
	GoDir           string
}

//GoRoot returns go root
func (s *Snapshot) GoRoot() string {
	return path.Join(s.GoDir, "go"+s.GoBuild.Version, "go")
}

//Env returns go env
func (s *Snapshot) Env() []string {
	goRootEnv := "GOROOT=" + path.Join(s.GoDir, "go"+s.GoBuild.Version, "go")
	homeEmv := "HOME=" + s.HomeURL()
	pathEnv := "PATH=/usr/bin:/usr/local/bin:/bin:/sbin:/usr/sbin"
	return []string{goRootEnv, homeEmv, pathEnv}
}

//BasePluginURL returns base plugin url
func (s *Snapshot) BasePluginURL() string {
	return path.Join(s.BaseDir, "plugin")
}

func (s *Snapshot) PluginMainURL() string {
	result := s.BasePluginURL()
	mainPath := s.Spec.MainPath

	if mainPath != "" {
		result = path.Join(result, mainPath)
	}

	return result
}

//HomeURL returns home dir
func (s *Snapshot) HomeURL() string {
	return path.Join(s.TempDir, "home")
}

//AppendMod append mod file
func (s *Snapshot) AppendMod(file *modfile.File) {
	s.ModFiles = append(s.ModFiles, file)
	s.setModFile(file)
}

func (s *Snapshot) setModFile(file *modfile.File) {
	if s.ModFile != nil || file.Module == nil {
		return
	}
	if s.Spec.ModPath == "" || s.Spec.ModPath == file.Module.Mod.Path {
		s.ModFile = file
		s.BuildModPath = file.Module.Mod.Path + "_t" + strconv.Itoa(int(time.Now().UnixMilli()))
		s.Spec.ModPath = file.Module.Mod.Path
	}
}

//AppendMain append main files
func (s *Snapshot) AppendMain(loc string) {
	s.Mains = append(s.Mains, loc)
	s.setPluginBuildPath(loc)
}

func (s *Snapshot) setPluginBuildPath(loc string) {
	if s.PluginBuildPath != "" {
		return
	}
	if s.Spec.MainPath == "" {
		s.PluginBuildPath = path.Join(s.BasePluginURL(), path.Dir(loc))
		return
	}

	if strings.Contains(loc, s.Spec.MainPath) {
		s.PluginBuildPath = path.Join(s.BasePluginURL(), path.Dir(loc))
	}
}

func (s *Snapshot) replaceDependencies(source []byte) ([]byte, error) {
	if s.Spec.ModPath == "" {
		return source, fmt.Errorf("mod path was empty")
	}

	if !bytes.Contains(source, []byte(s.Spec.ModPath)) {
		return source, nil
	}
	//TODO support for mode locally supplied dependencies with repalce
	return bytes.ReplaceAll(source, []byte(s.Spec.ModPath), []byte(s.BuildModPath)), nil
}

func (s *Snapshot) buildCmdArgs(buildSpec *build.Build) (string, []string) {
	args := []string{
		"build",
	}
	if s.buildMode != "exec" {
		args = append(args, "-buildmode=plugin")
	}
	if len(s.Spec.BuildArgs) > 0 {
		for _, arg := range s.Spec.BuildArgs {
			args = append(args, Args(arg).Elements()...)
		}
	}
	if s.GoBuild.LdFlags != "" {
		args = append(args, `-ldflags="`+s.GoBuild.LdFlags+`"`)
	}

	args = append(args,
		"-o",
		s.PluginDestPath,
	)

	mainPath := s.Spec.MainPath
	if s.PluginBuildPath == "" && mainPath != "" {
		args = append(args, mainPath)
	}

	return path.Join(s.GoRoot(), "bin", "go"), args
}

//NewSnapshot creates a snapshot
func NewSnapshot(buildMode string, plugin *build.Spec, goBuild build.GoBuild) *Snapshot {
	ret := &Snapshot{buildMode: buildMode, Spec: plugin, GoBuild: goBuild, Created: time.Now()}
	ret.TempDir = os.TempDir()
	ret.BaseDir = path.Join(ret.TempDir, strconv.Itoa(int(ret.Created.UnixMicro())))
	_ = os.MkdirAll(ret.BaseDir, defaultDirPermission)
	ret.GoDir = path.Join(ret.TempDir, "go")
	_ = os.MkdirAll(ret.GoDir, defaultDirPermission)
	_ = os.MkdirAll(ret.HomeURL(), defaultDirPermission)
	ret.PluginDestPath = path.Join(ret.BaseDir, "main.so")

	return ret
}
