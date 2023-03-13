package builder

import (
	"regexp"
	"strconv"
	"strings"
)

//Args represents cli arguments
type Args string

var argElements = regexp.MustCompile(`([-a-zA-Z]+)|(".*?[^\\]")|("")`)

//Elements returns args elements
func (a Args) Elements() []string {
	pluginArgs := argElements.FindAllString(string(a), -1)
	for i, arg := range pluginArgs {
		if !strings.HasPrefix(arg, `"`) || !strings.HasSuffix(arg, `"`) {
			continue
		}
		var err error
		pluginArgs[i], err = strconv.Unquote(arg)
		if err != nil {
			pluginArgs[i] = arg
		}
	}
	return pluginArgs
}
