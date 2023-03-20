package build

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/viant/afs"
	"github.com/viant/afs/file"
	"github.com/viant/afs/url"
	"io"
	"strings"
)

//Module represents plugin
type Module struct {
	Data []byte
	Mode string
	Info
}

//Store stores plugin in supplied dest
func (p *Module) Store(ctx context.Context, fs afs.Service, location string) error {
	data := p.Data

	if p.Mode == "exec" {
		if p.Name == "" {
			p.Name = "main"
		}
		dest := url.Join(location, p.Name)
		return fs.Upload(ctx, dest, file.DefaultFileOsMode, bytes.NewReader(data))
	}

	var err error
	dest := url.Join(location, p.Runtime.PluginName(p.Name))
	infoDest := strings.Replace(dest, ".so", ".info", 1)
	if p.Compression == "gzip" {
		if data, err = compressWithGzip(data); err != nil {
			return err
		}
		dest += ".gz"
	}
	if err = fs.Upload(ctx, dest, file.DefaultFileOsMode, bytes.NewReader(data)); err != nil {
		return err
	}
	if info, err := json.Marshal(p.Info); err == nil {
		if err := fs.Upload(ctx, infoDest, file.DefaultFileOsMode, bytes.NewReader(info)); err != nil {
			return err
		}
	}
	return nil
}

func compressWithGzip(data []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := gzip.NewWriter(buf)
	io.Copy(writer, bytes.NewReader(data))
	if err := writer.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), writer.Close()
}
