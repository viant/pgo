package build

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/viant/afs"
	"github.com/viant/afs/file"
	"github.com/viant/afs/url"
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
	if p.Mode == "exec" {
		if p.Name == "" {
			p.Name = "main"
		}
		dest := url.Join(location, p.Name)
		return fs.Upload(ctx, dest, file.DefaultFileOsMode, bytes.NewReader(p.Data))
	}

	dest := url.Join(location, p.Runtime.PluginName(p.Name))
	if err := fs.Upload(ctx, dest, file.DefaultFileOsMode, bytes.NewReader(p.Data)); err != nil {
		return err
	}
	if info, err := json.Marshal(p.Info); err == nil {
		if err := fs.Upload(ctx, strings.Replace(dest, ".so", ".info", 1), file.DefaultFileOsMode, bytes.NewReader(info)); err != nil {
			return err
		}
	}
	return nil
}
