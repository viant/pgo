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

//Plugin represents plugin
type Plugin struct {
	Data []byte
	Info
}

//Store stores plugin in supplied dest
func (p *Plugin) Store(ctx context.Context, fs afs.Service, location string) error {
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
