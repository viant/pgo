package build

import (
	"context"
	"fmt"
	"github.com/viant/afs"
	"testing"
)

func TestSo(t *testing.T) {

	s := &Source{}
	s.URL = "/Users/awitas/go/src/github.vianttech.com/adelphic/datlydev/.build/datly/"
	err := s.Pack(context.Background(), afs.New())
	fmt.Printf("%v\n", err)
}
