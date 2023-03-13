package build

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInfo_Decode(t *testing.T) {
	var testCases = []struct {
		description string
		URL         string
		expect      *Info
	}{
		{
			description: "encoded URL",
			URL:         "file:///tmp/main_1_17_6_darwin_amd64.so",
			expect: &Info{Runtime: Runtime{
				Arch:    "amd64",
				Os:      "darwin",
				Version: "1.17.6",
			}},
		},
	}

	for _, testCase := range testCases {
		info := &Info{}
		info.DecodeURL(testCase.URL)
		assert.EqualValues(t, testCase.expect, info, testCase.description)

	}

}
