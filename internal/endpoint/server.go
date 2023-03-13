package endpoint

import (
	"context"
	"fmt"
	"github.com/viant/pgo/internal"
	"log"
	"net/http"
	"os"
	"os/signal"
)

//Server represents a server
type Server struct {
	server http.Server
	port   int
	router *Router
}

//Start starts a server
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc(internal.BuildURI, func(writer http.ResponseWriter, request *http.Request) {
		if err := s.router.buildPlugin(writer, request); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc(internal.StatusURI, func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("up"))
	})

	s.server.Handler = mux
	s.server.Addr = fmt.Sprintf(":%v", s.port)
	s.shutdownOnInterrupt()
	return s.server.ListenAndServe()
}

//Shutdown stops server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

//ShutdownOnInterrupt
func (s *Server) shutdownOnInterrupt() {
	closed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		// We received an interrupt signal, shut down.
		if err := s.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(closed)
	}()
}

//NewServer creates a build server
func NewServer(port int, router *Router) *Server {
	return &Server{
		server: http.Server{},
		port:   port,
		router: router,
	}
}
