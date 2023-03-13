package endpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/viant/pgo/build"
	"github.com/viant/pgo/internal/builder"
	"io/ioutil"
	"net/http"
)

//Router represents a router
type Router struct {
	service *builder.Service
}

func (s *Router) buildPlugin(writer http.ResponseWriter, request *http.Request) error {
	if request.Method != http.MethodPost {
		return fmt.Errorf("unsupported method, expected POST, but had: %v", request.Method)
	}
	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return err
	}
	buildReq := &build.Build{}
	if err = json.Unmarshal(data, buildReq); err != nil {
		return err
	}
	response, err := s.service.Build(context.Background(), buildReq, build.WithLogger(nil))
	if err != nil {
		return err
	}
	if data, err = json.Marshal(response); err != nil {
		return err
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	return err
}

//NewRouter creates a router
func NewRouter(service *builder.Service) *Router {
	return &Router{service: service}
}
