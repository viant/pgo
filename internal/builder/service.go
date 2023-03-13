package builder

import (
	"bytes"
	"context"
	"fmt"
	"github.com/viant/afs"
	"github.com/viant/pgo/build"
	"golang.org/x/mod/modfile"
	"io"
	"io/ioutil"
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
func (s *Service) Build(ctx context.Context, buildSpec *build.Build, opts ...build.Option) (*build.Plugin, error) {
	for _, opt := range opts {
		opt(buildSpec)
	}
	buildSpec.Init()
	err := buildSpec.Validate()
	if err != nil {
		return nil, err
	}
	if err := s.cfg.Runtime.ValidateOsAndArch(&buildSpec.Go.Runtime); err != nil {
		return s.delegateBuildOrFail(ctx, buildSpec, err)
	}
	if len(buildSpec.Source.Data) == 0 {
		if err = buildSpec.Source.Pack(ctx, s.fs); err != nil {
			return nil, err
		}
	}
	snapshot := NewSnapshot(buildSpec.Plugin, buildSpec.Go)
	if err := s.ensureGo(ctx, snapshot, buildSpec.Go.Version, buildSpec.Logf); err != nil {
		return nil, err
	}
	if err = buildSpec.Source.Unpack(ctx, s.fs, snapshot.BasePluginURL(),
		func(mod *modfile.File) {
			snapshot.AppendMod(mod)
		},
		func(parent string, info os.FileInfo, reader io.ReadCloser) (os.FileInfo, io.ReadCloser, error) {
			ext := path.Ext(info.Name())
			switch ext {
			case ".go", ".mod", ".sum":
				return s.processSource(reader, parent, info, snapshot)
			}
			return info, reader, nil
		}); err != nil {
		return nil, err
	}

	if snapshot.PluginBuildPath == "" {
		buildSpec.Logf("failed to detect plugin main package\n")
		return nil, fmt.Errorf("failed to detect plugin main package")
	}
	if err = s.buildPlugin(snapshot, buildSpec); err != nil {
		return nil, err
	}
	pluginData, err := s.fs.DownloadWithURL(ctx, snapshot.PluginDestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to locate plugin: %v", err)
	}
	res := &build.Plugin{
		Data: pluginData,
		Info: build.Info{
			Scn:     build.AsScn(snapshot.Created),
			Runtime: buildSpec.Go.Runtime,
		},
	}
	return res, nil
}

func (s *Service) buildPlugin(snapshot *Snapshot, buildSpec *build.Build) error {
	cmd, args := snapshot.buildCmdArgs()
	command := exec.Command(cmd, args...)
	command.Dir = snapshot.PluginBuildPath
	command.Env = snapshot.Env()
	buildSpec.Logf("building plugin at %v: %v", command.Dir, command.String())
	output, err := command.CombinedOutput()
	if err != nil {
		buildSpec.Logf("couldn't generate plugin due to the: %w at: %s\n\tstdin: %s\n\tstdount: %s", err, command.Dir, command.String(), output)
		return fmt.Errorf("couldn't generate plugin due to the: %w at: %s\n\tstdin: %s\n\tstdount: %s", err, command.Dir, command.String(), output)
	}
	return nil
}

var mainFragment = []byte("package main")

func (s *Service) processSource(reader io.ReadCloser, parent string, info os.FileInfo, snapshot *Snapshot) (os.FileInfo, io.ReadCloser, error) {
	source, err := ioutil.ReadAll(reader)
	if err != nil {
		return info, reader, err
	}
	_ = reader.Close()
	source, err = snapshot.replaceDependencies(source)
	if err != nil {
		return info, reader, err
	}
	if bytes.Contains(source, mainFragment) {
		snapshot.AppendMain(path.Join(parent, info.Name()))
	}
	return info, ioutil.NopCloser(bytes.NewReader(source)), nil
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

func (s *Service) delegateBuildOrFail(ctx context.Context, spec *build.Build, err error) (*build.Plugin, error) {
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
