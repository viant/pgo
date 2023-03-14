package build

import (
	"strconv"
	"time"
)

const ScnLayout = "20060102150405"

//AsScn converts time to sequence change number
func AsScn(ts time.Time) int {
	out := ts.In(time.UTC).Format(ScnLayout)
	res, _ := strconv.Atoi(out)
	return res
}
