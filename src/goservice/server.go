package goservice

import (
	"os"
)

type Server struct {
	services API
	endpoints []Endpoint

	stopper chan os.Signal
}

func NewServer(services API) *Server {
	return &Server {
		services: services,
		endpoints: make([]Endpoint, 0),
	}
}

func (server *Server) AddEndpoint(endpoint Endpoint) {
	server.endpoints = append(server.endpoints, endpoint)
}

func (server *Server) Start() {
	for _, endpoint := range server.endpoints {
		endpoint.Start()
	}
}

func (server *Server) Stop() {
	for _, endpoint := range server.endpoints {
		endpoint.Stop()
	}
}


// ------------------------------------------
// Context API
// ------------------------------------------

func (server *Server) API() API {
	return server.services
}