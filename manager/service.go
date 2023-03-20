package manager

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/viant/afs"
	"github.com/viant/afs/file"
	"github.com/viant/afs/url"
	"github.com/viant/pgo/build"
	"io"
	"os"
	"path"
	"plugin"
	"strconv"
	"strings"
	"sync"
)

//Service re
type Service struct {
	scn     build.SequenceChangeNumber
	fs      afs.Service
	runtime build.Runtime
	mux     sync.RWMutex
	info    map[string]*build.Info
}

//OpenWithInfoURL open plugin based on the plugin info URL
func (s *Service) OpenWithInfoURL(ctx context.Context, URL string) (*build.Info, *plugin.Plugin, error) {
	info, err := s.loadInfo(ctx, URL)
	if err != nil {
		return nil, nil, err
	}
	URL = strings.Replace(URL, ".pinf", ".so", 1)
	if info.Compression == "gzip" {
		URL += ".gz"
	}
	provider, err := s.Open(ctx, URL)
	return info, provider, err
}

//Open open plugin, if plugin is latest server SCN
func (s *Service) Open(ctx context.Context, URL string) (*plugin.Plugin, error) {
	isCompressed := strings.HasSuffix(URL, ".gz")
	if isCompressed {
		URL = URL[:len(URL)-3]
	}
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
	if schema != file.Scheme || info.Compression != "" || isCompressed {
		location = path.Join(os.TempDir(), strconv.Itoa(int(info.Scn)))
		_ = os.MkdirAll(location, 0o744)
		location = path.Join(location, path.Base(URL))

		if isCompressed {
			data, err := s.fs.DownloadWithURL(ctx, URL+".gz")
			if err != nil {
				return nil, fmt.Errorf("failed to load %w; %v", err, URL)
			}
			if data, err = gunzip(err, data); err != nil {
				return nil, fmt.Errorf("failed to unzip %w; %v", err, URL)
			}
			if err = s.fs.Upload(ctx, location, file.DefaultFileOsMode, bytes.NewReader(data)); err != nil {
				return nil, fmt.Errorf("failed upload: %w;  %v to %v", err, URL, location)
			}
		} else if err := s.fs.Copy(ctx, URL, location); err != nil {
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

func gunzip(err error, data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	reader.Close()
	buf := new(bytes.Buffer)
	io.Copy(buf, reader)
	data = buf.Bytes()
	return data, nil
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
	infoData, err := s.fs.DownloadWithURL(ctx, strings.Replace(URL, ".so", ".pinf", 1))
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

	result.Scn = build.NewSequenceChangeNumber(obj.ModTime())
	return result, nil
}

//New represents a manager
func New(scn build.SequenceChangeNumber) *Service {
	ret := &Service{scn: scn, info: map[string]*build.Info{}, fs: afs.New()}
	ret.runtime.Init()
	return ret
}
