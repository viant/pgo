package main

import (
	"github.com/viant/pgo/internal/builder"
	"github.com/viant/pgo/internal/endpoint"
	"os"
	"strconv"
)

func main() {
	aBuilder := builder.New(&builder.Config{})
	router := endpoint.NewRouter(aBuilder)
	port := 8089
	if portLiteral := os.Getenv("PORT"); portLiteral != "" {
		port, _ = strconv.Atoi(portLiteral)
	}
	srv := endpoint.NewServer(port, router)
	srv.Start()
}
