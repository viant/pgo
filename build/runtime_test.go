package build

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRuntime_DetectVersion(t *testing.T) {
	runtime := &Runtime{}
	runtime.DetectVersion()
	assert.Truef(t, runtime.Version != "", "expected version")
}
