package manager

import "errors"

var (
	errScnToOld = errors.New("plugin scn too old")
)

//IsPluginToOld returns true if plugin it too old
func IsPluginToOld(err error) bool {
	return errScnToOld == err
}
