package goservice

import (
	"os"
)

type Server struct {
	services API

	sessionCreator SessionCreator
	endpoints []Endpoint

	stopper chan os.Signal
}

func NewServer(services API, sessionCreator SessionCreator) *Server {
	return &Server {
		services: services,
		sessionCreator: sessionCreator,
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

func (server *Server) CreateSession(endpoint Endpoint) Session {
	return server.sessionCreator(endpoint)
}
