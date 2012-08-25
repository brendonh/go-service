package goservice

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"strconv"
	"encoding/json"
)

type SessionResolver func(*http.Request, ServerContext) (Session, error)

type HttpRpcEndpoint struct {
	Address string
	listener net.Listener
	context ServerContext
	resolver SessionResolver
}


func NewHttpRpcEndpoint(address string, context ServerContext, resolver SessionResolver) Endpoint {
	if resolver == nil {
		resolver = DefaultSessionResolver
	}
	return &HttpRpcEndpoint{
		Address: address,
		context: context,
		resolver: resolver,
	}
}

func DefaultSessionResolver(req *http.Request, context ServerContext) (Session, error) {
	// XXX TODO: Session tracking
	return NewBasicSession(), nil
}

func (endpoint *HttpRpcEndpoint) Start() bool {
	if endpoint.listener != nil {
		return false
	}

	listener, error := net.Listen("tcp", endpoint.Address)
	if error != nil {
		fmt.Printf("Error starting HTTP RPC endpoint: %v\n", error)
		return false
	}

	endpoint.listener = listener

	mux := http.NewServeMux()
	mux.HandleFunc("/favicon.ico", http.NotFound)
	mux.Handle("/", endpoint)
	go http.Serve(listener, mux)

	fmt.Printf("HTTP endpoint started at %s\n", endpoint.Address)

	return true
}


func (endpoint *HttpRpcEndpoint) Stop() bool {
	if endpoint.listener == nil {
		return true
	}

	if error := endpoint.listener.Close(); error != nil {
		fmt.Printf("Error stopping HTTP RPC endpoint: %v\n", error)
		return false
	}

	endpoint.listener = nil
	return true
}


func (endpoint *HttpRpcEndpoint) ServeHTTP(response http.ResponseWriter, req *http.Request) {
	bits := strings.SplitN(req.URL.Path[1:], "/", 2)

	if len(bits) != 2 {
		http.NotFound(response, req)
		return
	}

	req.ParseForm()

	var form = make(APIData)
	for k, v := range req.Form {
		form[k] = v[0]
	}

	var session, err = endpoint.resolver(req, endpoint.context)
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

	response.Header().Add("Content-Type", "text/plain")

	jsonReply, _ := json.Marshal(Response(ok, errors, resp))
	response.Header().Add("Content-Length", strconv.Itoa(len(jsonReply)))

	response.Write(jsonReply)
}