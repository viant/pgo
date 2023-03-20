package manager

import "errors"

var (
	errScnToOld = errors.New("plugin scn is outdated")
)

//IsPluginOutdated returns true if plugin it too old
func IsPluginOutdated(err error) bool {
	return errScnToOld == err
}
