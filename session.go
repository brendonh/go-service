package goservice

import (
	"sync"
	"fmt"

	"code.google.com/p/go-uuid/uuid"
)


func NewBasicSession(endpoint Endpoint) *BasicSession {
	return &BasicSession {
		id: uuid.New(),
		user: nil,
		Mutex: new(sync.Mutex),
		endpoint: endpoint,
	}
}


func BasicSessionCreator(endpoint Endpoint) Session {
	return NewBasicSession(endpoint)
}

// ------------------------------------------
// Session API
// ------------------------------------------

func (session *BasicSession) ID() string {
	return session.id;
}

func (session *BasicSession) User() User {
	return session.user;
}

func (session *BasicSession) SetUser(user User) {
	fmt.Printf("Session login: %s (%s)\n", user.DisplayName(), session.id)
	session.user = user
}

func (session *BasicSession) Send(msg []byte) {
	fmt.Printf("Session send: %s: %v\n", session.id, msg)
}