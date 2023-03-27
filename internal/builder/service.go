package builder

import (
	"bytes"
	"context"
	"fmt"
	"github.com/viant/afs"
	"github.com/viant/pgo/build"
	"golang.org/x/mod/modfile"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

//Service represents builder service
type Service struct {
	cfg    *Config
	fs     afs.Service
	logger func(template string, args ...interface{})
}

//Build builds plugin
func (s *Service) Build(ctx context.Context, buildSpec *build.Build, opts ...build.Option) (*build.Module, error) {
	for _, opt := range opts {
		opt(buildSpec)
	}
	buildSpec.Init()
	err := buildSpec.Validate()
	if err != nil {
		return nil, err
	}
	if err = s.packSourceIfNeeded(ctx, buildSpec); err != nil {
		return nil, err
	}
	if err := s.cfg.Runtime.ValidateOsAndArch(&buildSpec.Go.Runtime); err != nil || buildSpec.Go.EnsureTheSameOs {
		return s.delegateBuildOrFail(ctx, buildSpec, err)
	}

	snapshot := NewSnapshot(buildSpec.Name, buildSpec.Mode, &buildSpec.Spec, buildSpec.Go)
	if err := s.ensureGo(ctx, snapshot, buildSpec.Go.Version, buildSpec.Logf); err != nil {
		return nil, err
	}
	if err = s.unpackDependencies(ctx, buildSpec, snapshot); err != nil {
		return nil, err
	}
	if err = buildSpec.Source.Unpack(ctx, s.fs, snapshot.BaseModuleURL(),
		func(mod *modfile.File) {
			snapshot.AppendMod(mod)
		},
		func(parent string, info os.FileInfo, reader io.ReadCloser) (os.FileInfo, io.ReadCloser, error) {
			if info.Name() == "go.mod" {
				if reader, err = s.replaceLocalDependencies(reader, snapshot, buildSpec); err != nil {
					return nil, nil, err
				}
			}
			ext := path.Ext(info.Name())
			switch ext {
			case ".go", ".sum", ".mod":
				return s.processSource(reader, parent, info, snapshot, snapshot.buildMode != "exec")
			}
			return info, reader, nil
		}); err != nil {
		return nil, err
	}

	if snapshot.buildMode != "exec" {
		buildSpec.Logf("build module: %s from %s\n", snapshot.BuildModPath, snapshot.Spec.ModPath)
	}
	if err = s.build(snapshot, buildSpec); err != nil {
		return nil, err
	}

	data, err := s.fs.DownloadWithURL(ctx, snapshot.ModuleDestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to locate plugin: %v", err)
	}

	res := &build.Module{
		Mode: snapshot.buildMode,
		Data: data,
		Info: build.Info{
			Scn:     build.NewSequenceChangeNumber(snapshot.Created),
			Runtime: buildSpec.Go.Runtime,
			Name:    buildSpec.Name,
		},
	}
	return res, nil
}

func (s *Service) replaceLocalDependencies(reader io.ReadCloser, snapshot *Snapshot, spec *build.Build) (io.ReadCloser, error) {
	if len(snapshot.Dependencies) == 0 {
		return reader, nil
	}
	replace := snapshot.ModFile.Replace
	if len(replace) == 0 {
		return reader, nil
	}
	source, err := io.ReadAll(reader)
	if err != nil {
		return reader, err
	}
	_ = reader.Close()
	sourceCode := string(source)
	for _, item := range replace {
		dep := snapshot.MatchDependency(item.Old)
		if dep == nil {
			spec.Logf("failed to match replace: %v\n", item.Old)
			continue
		}
		elem := []string{dep.BaseURL}
		if dep.Parent != "" {
			elem = append(elem, dep.Parent)
		}
		newPath := path.Join(elem...)
		spec.Logf("replacing mode dep: %v %v\n", item.New.Path, newPath)
		sourceCode = strings.ReplaceAll(sourceCode, item.New.Path, newPath)
	}
	return io.NopCloser(strings.NewReader(sourceCode)), nil
}

//replace github.com/viant/xdatly/types/custom => /Users/awitas/go/src/github.vianttech.com/adelphic/datlydev/pkg

func (s *Service) unpackDependencies(ctx context.Context, buildSpec *build.Build, snapshot *Snapshot) error {
	if len(buildSpec.LocalDep) > 0 {
		for i, localDep := range buildSpec.LocalDep {
			dep := &Dependency{}
			snapshot.Dependencies = append(snapshot.Dependencies, dep)
			destURL := snapshot.BaseDependencyURL(i)
			if err := localDep.Unpack(ctx, s.fs, destURL,
				func(mod *modfile.File) {
					buildSpec.Logf("detected go mod: %v with: %v\n", mod.Module.Mod.String(), destURL)
					dep.BaseURL = destURL
					if dep.Mod != nil {
						return
					}
					dep.Mod = mod
				}, func(parent string, info os.FileInfo, reader io.ReadCloser) (os.FileInfo, io.ReadCloser, error) {
					if info.Name() == "go.mod" && !!dep.ParentSet {
						buildSpec.Logf("detected go mod path: (%v) %v\n", parent, info.Name())
						dep.ParentSet = true
						dep.Parent = parent
					}
					return info, reader, nil
				}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) packSourceIfNeeded(ctx context.Context, buildSpec *build.Build) error {
	if len(buildSpec.LocalDep) > 0 {
		for _, dep := range buildSpec.LocalDep {
			if len(dep.Data) == 0 {
				if err := dep.Pack(ctx, s.fs); err != nil {
					return err
				}
			}
		}
	}
	if len(buildSpec.Source.Data) == 0 {
		if err := buildSpec.Source.Pack(ctx, s.fs); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) build(snapshot *Snapshot, buildSpec *build.Build) error {
	cmd, args := snapshot.buildCmdArgs()
	command := exec.Command(cmd, args...)
	command.Dir = snapshot.ModuleBuildPath
	command.Env = appendEnv(buildSpec.Go.Env, snapshot.Env())
	buildSpec.Logf("building %v module at %v: %v", snapshot.buildMode, command.Dir, command.String())
	output, err := command.CombinedOutput()
	if err != nil {
		buildSpec.Logf("couldn't generate module due to the: %w at: %s\n\tstdin: %s\n\tstdount: %s", err, command.Dir, command.String(), output)
		return fmt.Errorf("couldn't generate module due to the: %w at: %s\n\tstdin: %s\n\tstdount: %s", err, command.Dir, command.String(), output)
	}
	return nil
}

func appendEnv(pairs map[string]string, env []string) []string {
	if len(pairs) > 0 {
		for k, v := range pairs {
			env = append(env, fmt.Sprintf("%v=%v", k, v))
		}
	}
	return env
}

var mainFragment = []byte("package main")

func (s *Service) processSource(reader io.ReadCloser, parent string, info os.FileInfo, snapshot *Snapshot, replace bool) (os.FileInfo, io.ReadCloser, error) {
	source, err := io.ReadAll(reader)
	if err != nil {
		return info, reader, err
	}
	_ = reader.Close()
	if replace {
		source, err = snapshot.replaceDependencies(source)
		if err != nil {
			return info, reader, err
		}
	}
	if bytes.Contains(source, mainFragment) {
		snapshot.AppendMain(path.Join(parent, info.Name()))
	}
	return info, io.NopCloser(bytes.NewReader(source)), nil
}

var goDownloadURL = "https://dl.google.com/go/go%v.%v-%v.tar.gz"

func (s *Service) ensureGo(ctx context.Context, snapshot *Snapshot, version string, logf func(format string, args ...interface{})) error {
	verLocation := path.Join(snapshot.GoDir, "go"+version)
	ok, _ := s.fs.Exists(ctx, path.Join(verLocation, "go"))
	logf("checking binary[%v]: %v\n", ok, verLocation)
	if ok {
		return nil
	}
	if err := os.MkdirAll(verLocation, defaultDirPermission); err != nil {
		return fmt.Errorf("failed to crate %v %v", verLocation, err)
	}
	URL := fmt.Sprintf(goDownloadURL, version, s.cfg.Runtime.Os, s.cfg.Runtime.Arch)
	URL = strings.Replace(URL, "://", ":", 1) + "/tar://"
	logf("installing go %v %v %v\n", version, s.cfg.Runtime.Os, s.cfg.Runtime.Arch)
	err := s.fs.Copy(ctx, URL, verLocation)
	if err != nil {
		logf("failed to install go %v\n", err)
	}
	return err
}

func (s *Service) delegateBuildOrFail(ctx context.Context, spec *build.Build, err error) (*build.Module, error) {
	spec.Go.EnsureTheSameOs = false //do not propagate that flag down otherwise infinitive loop
	delegation := s.cfg.delegations.Match(&spec.Go.Runtime)
	if delegation == nil {
		return nil, err
	}
	if err := s.ensureDocker(delegation, spec); err != nil {
		return nil, err
	}
	aClient := NewClient(delegation.baseURL())
	return aClient.Build(ctx, spec)
}

func (s *Service) ensureDocker(delegation *Delegation, spec *build.Build) error {
	aClient := NewClient(delegation.baseURL())
	if aClient.IsUp() {
		spec.Logf("%v is up\n", delegation.Name)
		return nil
	}
	return s.runDocker(delegation, spec)
}

//New creates a service
func New(cfg *Config, opts ...Option) *Service {
	cfg.Runtime.Init()
	for _, opt := range opts {
		opt(cfg)
	}
	return &Service{fs: afs.New(), cfg: cfg}
}
