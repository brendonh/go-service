package goservice

import (
	"os"
	"log"
)

type Server struct {
	services API

	sessionCreator SessionCreator
	endpoints []Endpoint
	logger *log.Logger

	stopper chan os.Signal
}

func NewServer(services API, sessionCreator SessionCreator) *Server {
	return &Server {
		services: services,
		sessionCreator: sessionCreator,
		endpoints: make([]Endpoint, 0),
		logger: log.New(os.Stdout, "", log.LstdFlags),
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

func (server *Server) CreateSession(conn SessionConnection) Session {
	return server.sessionCreator(conn)
}

func (server *Server) Log(format string, args... interface{}) {
	server.LogPrefix("Server", format, args...)
}

func (server *Server) LogPrefix(prefix string, format string, args... interface{}) {
	server.logger.Printf("[ %-20s ] " + format + "\n", append([]interface{} { prefix }, args...)...)
}