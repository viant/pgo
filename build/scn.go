package build

import (
	"fmt"
	"strconv"
	"time"
)

const ScnLayout = "20060102150405"

type SequenceChangeNumber int

//NewSequenceChangeNumber creates a time based sequence change number
func NewSequenceChangeNumber(ts time.Time) SequenceChangeNumber {
	out := ts.In(time.UTC).Format(ScnLayout)
	res, _ := strconv.Atoi(out)
	return SequenceChangeNumber(res)
}

//AsTime converts
func (s SequenceChangeNumber) AsTime() (time.Time, error) {
	return time.ParseInLocation(ScnLayout, fmt.Sprintf("%v", int(s)), time.UTC)
}
