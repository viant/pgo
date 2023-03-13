package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/viant/afs"
	"github.com/viant/afs/file"
	"github.com/viant/afs/url"
	"github.com/viant/pgo/build"
	"os"
	"path"
	"plugin"
	"strconv"
	"strings"
	"sync"
)

//Service re
type Service struct {
	scn     int
	fs      afs.Service
	runtime build.Runtime
	mux     sync.RWMutex
	info    map[string]*build.Info
}

//Open open plugin, if plugin is latest server SCN
func (s *Service) Open(ctx context.Context, URL string) (*plugin.Plugin, error) {
	prev := s.getPluginInfo(URL)
	info, err := s.loadInfo(ctx, URL)
	if err != nil {
		return nil, err
	}
	if info.Scn < s.scn {
		return nil, errScnToOld
	}
	if prev != nil && prev.Scn >= info.Scn {
		return nil, errScnToOld
	}
	if err := s.runtime.Validate(&info.Runtime); err != nil {
		return nil, err
	}
	schema := url.Scheme(URL, file.Scheme)
	location := url.Path(URL)
	if schema != file.Scheme {
		location = path.Join(os.TempDir(), strconv.Itoa(info.Scn))
		_ = os.MkdirAll(location, 0o744)
		location = path.Join(location, path.Base(URL))
		if err := s.fs.Copy(ctx, URL, location); err != nil {
			return nil, fmt.Errorf("failed to copy: %w;  %v to %v", err, URL, location)
		}
	}
	aPlugin, err := plugin.Open(location)
	if err != nil {
		return nil, err
	}
	s.setPluginInfo(URL, info)
	return aPlugin, nil
}

func (s *Service) getPluginInfo(URL string) *build.Info {
	s.mux.RLock()
	ret := s.info[URL]
	s.mux.RUnlock()
	return ret
}

func (s *Service) setPluginInfo(URL string, info *build.Info) {
	s.mux.Lock()
	s.info[URL] = info
	s.mux.Unlock()
}

func (s *Service) loadInfo(ctx context.Context, URL string) (*build.Info, error) {
	infoData, err := s.fs.DownloadWithURL(ctx, strings.Replace(URL, ".so", ".info", 1))
	if err != nil {
		return s.decodeURLInfo(ctx, URL)
	}
	info := &build.Info{}
	return info, json.Unmarshal(infoData, info)
}

func (s *Service) decodeURLInfo(ctx context.Context, URL string) (*build.Info, error) {
	var result = &build.Info{}
	result.DecodeURL(URL)
	obj, err := s.fs.Object(ctx, URL)
	if err != nil {
		return nil, err
	}

	result.Scn = build.AsScn(obj.ModTime())
	return result, nil
}

//New represents a manager
func New(scn int) *Service {
	ret := &Service{scn: scn, info: map[string]*build.Info{}, fs: afs.New()}
	ret.runtime.Init()
	return ret
}
