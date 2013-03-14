package goservice

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"strconv"
	"encoding/json"
)

type SessionResolver func(*http.Request, http.ResponseWriter, *HttpRpcEndpoint) (Session, error)

type HttpRpcEndpoint struct {
	Address string
	Mux *http.ServeMux
	listener net.Listener
	context ServerContext
	resolver SessionResolver
	stripLength int
	logPrefix string
}


type HttpRpcEndpointOptions struct {
	Resolver SessionResolver
	Static bool
	StaticPath string
	StaticUri string
	APIUri string
}

var defaultOptions = &HttpRpcEndpointOptions{
	Resolver: DefaultSessionResolver,
	Static: false,
	APIUri: "/",
}

func NewHttpRpcEndpoint(address string, context ServerContext, options *HttpRpcEndpointOptions) Endpoint {
	if options == nil {
		options = defaultOptions
	}

	var mux = http.NewServeMux()
	mux.HandleFunc("/favicon.ico", http.NotFound)

	if options.Static {
		mux.Handle(options.StaticUri, http.FileServer(http.Dir(options.StaticPath)))
	}

	var endpoint = &HttpRpcEndpoint{
		Mux: mux,
		Address: address,
		context: context,
		resolver: options.Resolver,
		stripLength: len(options.APIUri),
		logPrefix: "HTTP " + address,
	}
	mux.Handle(options.APIUri, endpoint)

	return endpoint
}


type HttpSessionConnection struct {
}

func (sessConn *HttpSessionConnection) Send(msg []byte) {
	fmt.Printf("HTTP session send: %v\n", msg)
}


func DefaultSessionResolver(req *http.Request, response http.ResponseWriter, endpoint *HttpRpcEndpoint) (Session, error) {
	// XXX TODO: Session tracking, sending
	var sender = &HttpSessionConnection{}
	return endpoint.context.CreateSession(sender), nil
}

func (endpoint *HttpRpcEndpoint) Start() bool {
	if endpoint.listener != nil {
		return false
	}

	listener, error := net.Listen("tcp", endpoint.Address)
	if error != nil {
		endpoint.Log("Error starting HTTP RPC endpoint: %v", error)
		return false
	}

	endpoint.listener = listener

	go http.Serve(listener, endpoint.Mux)

	endpoint.Log("HTTP endpoint started at %s", endpoint.Address)

	return true
}


func (endpoint *HttpRpcEndpoint) Stop() bool {
	if endpoint.listener == nil {
		return true
	}

	if error := endpoint.listener.Close(); error != nil {
		endpoint.Log("Error stopping HTTP RPC endpoint: %v", error)
		return false
	}

	endpoint.listener = nil
	return true
}

func (endpoint *HttpRpcEndpoint) Context() ServerContext { 
	return endpoint.context
}

func (endpoint *HttpRpcEndpoint) Log(fmt string, args... interface{}) {
	endpoint.context.LogPrefix(endpoint.logPrefix, fmt, args...)
}

func (endpoint *HttpRpcEndpoint) ServeHTTP(response http.ResponseWriter, req *http.Request) {
	bits := strings.SplitN(req.URL.Path[endpoint.stripLength:], "/", 2)

	if len(bits) != 2 {
		http.NotFound(response, req)
		return
	}

	req.ParseForm()

	var form = make(APIData)
	for k, v := range req.Form {
		form[k] = v[0]
	}

	var session, err = endpoint.resolver(req, response, endpoint)
	if err != nil {
		response.WriteHeader(400)
		response.Header().Add("Content-Type", "text/plain")
		response.Write([]byte(fmt.Sprintf("Session rejected: %s", err.Error())))
		return
	}

	ok, errors, resp := endpoint.context.API().HandleCall(
		bits[0], bits[1], form, session, endpoint.context)

	if errors != nil {
		response.WriteHeader(400)
	}

	response.Header().Add("Content-Type", "application/json")

	jsonReply, _ := json.Marshal(Response(ok, errors, resp))
	response.Header().Add("Content-Length", strconv.Itoa(len(jsonReply)))

	response.Write(jsonReply)
}