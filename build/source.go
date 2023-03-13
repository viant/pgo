package build

import (
	"bytes"
	"context"
	"fmt"
	"github.com/viant/afs"
	"github.com/viant/afs/file"
	"github.com/viant/afs/option"
	"golang.org/x/mod/modfile"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	transientPath    = "localhost/temp/source"
	transientZipPath = "localhost/"
)

type (
	//Source represent plugin module code source
	Source struct {
		Data []byte
		URL  string
	}
)

//Pack packs URL to data
func (s *Source) Pack(ctx context.Context, fs afs.Service) error {
	ext := path.Ext(s.URL)
	if strings.ToLower(ext) == ".zip" {
		data, err := fs.DownloadWithURL(ctx, s.URL)
		if err != nil {
			return err
		}
		s.Data = data
		return nil
	}
	transientURL := s.transientURL()
	destURL := s.transientZipURL(transientURL)
	if err := fs.Copy(ctx, s.URL, destURL); err != nil {
		return err
	}
	data, err := fs.DownloadWithURL(ctx, transientURL)
	if err != nil {
		return err
	}
	_ = fs.Delete(ctx, transientURL)
	s.Data = data
	return nil
}

//Unpack unpacks data into dest URL
func (s *Source) Unpack(ctx context.Context, fs afs.Service, destURL string, modHandler func(mod *modfile.File), modifier option.Modifier) error {
	if len(s.Data) == 0 {
		return fmt.Errorf("data was empty")
	}
	transientURL := s.transientURL()

	if err := fs.Upload(ctx, transientURL, file.DefaultFileOsMode, bytes.NewReader(s.Data)); err != nil {
		return err
	}
	transientZipURL := s.transientZipURL(transientURL)

	if err := fs.Walk(ctx, transientZipURL, func(ctx context.Context, baseURL string, parent string, info os.FileInfo, reader io.Reader) (toContinue bool, err error) {
		if info.Name() == "go.mod" {
			modContent, _ := io.ReadAll(reader)
			aMod, err := modfile.Parse(info.Name(), modContent, nil)
			if err != nil {
				return false, err
			}
			modHandler(aMod)
		}
		return true, nil
	}); err != nil {
		return err
	}

	src := option.Source{modifier}
	return fs.Copy(ctx, transientZipURL, destURL, &src)
}

func (s *Source) transientURL() string {
	return "mem://" + s.transientPath()
}

func (s *Source) transientPath() string {
	nanoTimestamp := int(time.Now().UnixMilli())
	return transientPath + strconv.Itoa(nanoTimestamp)
}

func (s *Source) transientZipURL(URL string) string {
	URL = strings.Replace(URL, "://", ":", 1)
	return URL + "/zip://" + transientZipPath
}
