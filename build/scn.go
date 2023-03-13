package build

import (
	"strconv"
	"time"
)

//AsScn converts time to sequence change number
func AsScn(ts time.Time) int {
	out := ts.In(time.UTC).Format("20060102150405")
	res, _ := strconv.Atoi(out)
	return res
}
