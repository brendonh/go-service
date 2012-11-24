package goservice

import (
	"sync"
)


// ------------------------------------------
// Server
// ------------------------------------------

type ServerContext interface {
	API() API
	CreateSession() Session
}

type Endpoint interface {
	Start() bool
	Stop() bool
}


// ------------------------------------------
// Users and sessions
// ------------------------------------------

type User interface {
	ID() string
	DisplayName() string
}

type Session interface {
	ID() string
	User() User
	SetUser(User)

	Lock()
	Unlock()
}

type SessionCreator func() Session

type BasicSession struct {
	id string
	user User
	*sync.Mutex
}


// ------------------------------------------
// API
// ------------------------------------------

const (
	IntArg = iota
	FloatArg
	StringArg
	NestedArg
    RawArg
)

type APIArg struct {
	Name string
	ArgType int
	Required bool
	Default interface{}
	Extra interface{}
}

type APIMethod struct {
	Name string
	ArgSpec []APIArg
	Handler APIHandler
}

type APIData map[string]interface{}

type APIHandler func(APIData, Session, ServerContext) (bool, APIData)


// ------------------------------------------
// Services
// ------------------------------------------

type APIService interface {
	Name() string
	AddMethod(string, []APIArg, APIHandler)
	FindMethod(string) *APIMethod
}

type API interface {
	AddService(APIService)
	HandleRequest(APIData, Session, ServerContext) APIData
	HandleCall(string, string, APIData, Session, ServerContext) (bool, []string, APIData)
}
