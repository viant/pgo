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

// Snapshot represent a plugin snapshot
type (
	Snapshot struct {
		Created   time.Time
		buildMode string
		Spec      *build.Spec
		build.GoBuild
		ModuleDestPath  string
		ModuleBuildPath string
		ModFile         *modfile.File
		BuildModPath    string
		ModFiles        []*modfile.File
		Mains           []string
		BaseDir         string
		TempDir         string
		GoDir           string
		goRoot          string
		GoPath          string
		Dependencies    []*Dependency
	}

	Dependency struct {
		Mod       *modfile.File
		Parent    string
		ParentSet bool
		BaseURL   string
		OriginURL string
	}
)

// GoRoot returns go root
func (s *Snapshot) GoRoot() string {
	if s.goRoot != "" {
		return s.goRoot
	}
	return path.Join(s.GoDir, "go"+s.GoBuild.Version, "go")
}

func (s *Snapshot) MatchDependency(match *modfile.Replace) *Dependency {
	if len(s.Dependencies) == 0 {
		return nil
	}
	for _, candidate := range s.Dependencies {
		if candidate.Mod == nil {
			continue
		}
		if candidate.Mod.Module.Mod.Path == match.Old.Path || match.New.Path == candidate.OriginURL {
			return candidate
		}
	}
	return nil
}

// Env returns go env
func (s *Snapshot) Env() []string {
	goRootEnv := "GOROOT=" + s.GoRoot()
	homeEmv := "HOME=" + s.HomeURL()
	goPath := s.GoPath
	if goPath == "" {
		goPath = path.Join(s.HomeURL(), "go")
	}
	goPathEnv := "GOPATH=" + goPath
	pathEnv := "PATH=/usr/bin:/usr/local/bin:/bin:/sbin:/usr/sbin"
	return []string{goRootEnv, homeEmv, pathEnv, goPathEnv}
}

// BaseModuleURL returns base plugin url
func (s *Snapshot) BaseModuleURL() string {
	return path.Join(s.BaseDir, "plugin")
}

// BaseDependencyURL returns baseDependency
func (s *Snapshot) BaseDependencyURL(index int) string {
	return path.Join(s.BaseDir, fmt.Sprintf("dep%v", index))
}

func (s *Snapshot) PluginMainURL() string {
	result := s.BaseModuleURL()
	mainPath := s.Spec.MainPath

	if mainPath != "" {
		result = path.Join(result, mainPath)
	}

	return result
}

// HomeURL returns home dir
func (s *Snapshot) HomeURL() string {
	return path.Join(s.TempDir, "home")
}

// AppendMod append mod file
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

// AppendMain append main files
func (s *Snapshot) AppendMain(loc string) {
	s.Mains = append(s.Mains, loc)
	s.setPluginBuildPath(loc)
}

func (s *Snapshot) setPluginBuildPath(loc string) {
	if s.ModuleBuildPath != "" {
		return
	}
	if path.IsAbs(s.Spec.MainPath) {
		s.ModuleBuildPath = s.Spec.MainPath
		return
	}
	if s.Spec.MainPath == "" {
		s.ModuleBuildPath = path.Join(s.BaseModuleURL(), path.Dir(loc))
		return
	}
	if strings.Contains(loc, s.Spec.MainPath) {
		fmt.Printf("matched mainPath: %v %v\n", s.Spec.MainPath, loc)
		s.ModuleBuildPath = path.Join(s.BaseModuleURL(), path.Dir(loc))
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

func (s *Snapshot) tidyCmdArgs() (string, []string) {
	return path.Join(s.GoRoot(), "bin", "go"), []string{
		"mod", "tidy",
	}
}

func (s *Snapshot) buildCmdArgs() (string, []string) {
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
		args = append(args, `-ldflags=`+s.GoBuild.LdFlags)
	}

	args = append(args,
		"-o",
		s.ModuleDestPath,
	)

	mainPath := s.Spec.MainPath
	if path.IsAbs(mainPath) && len(s.Mains) > 0 {
		s.Mains[0] = mainPath
	}
	if s.BuildModPath == "" && len(s.Mains) > 0 {
		args = append(args, s.Mains[0])
	} else if s.ModuleBuildPath == "" && mainPath != "" {
		args = append(args, mainPath)
	}

	return path.Join(s.GoRoot(), "bin", "go"), args
}

func (s *Snapshot) buildGoDownloadCmdArgs() (string, []string) {
	args := []string{
		"mod",
		"download",
	}
	return path.Join(s.GoRoot(), "bin", "go"), args
}

// NewSnapshot creates a snapshot
func NewSnapshot(name string, buildMode string, spec *build.Spec, goBuild build.GoBuild) *Snapshot {
	ret := &Snapshot{buildMode: buildMode, Spec: spec, GoBuild: goBuild, Created: time.Now()}
	ret.TempDir = os.TempDir()
	ret.BaseDir = path.Join(ret.TempDir, strconv.Itoa(int(ret.Created.UnixMicro())))
	_ = os.MkdirAll(ret.BaseDir, defaultDirPermission)

	if goBuild.Root != "" {
		ret.goRoot = goBuild.Root
	}
	if goBuild.Path != "" {
		ret.GoPath = goBuild.Path
	}
	ret.GoDir = path.Join(ret.TempDir, "go")
	_ = os.MkdirAll(ret.GoDir, defaultDirPermission)
	_ = os.MkdirAll(ret.HomeURL(), defaultDirPermission)
	if name == "" {
		name = "main"
	}
	ret.ModuleDestPath = path.Join(ret.BaseDir, name)

	return ret
}
