package goservice

import (
	"sync"
	"fmt"

	"code.google.com/p/go-uuid/uuid"
)


func NewSession() *Session {
	return &Session {
		id: uuid.New(),
		user: nil,
		Mutex: new(sync.Mutex),
	}
}


// ------------------------------------------
// Session API
// ------------------------------------------

func (session *Session) ID() string {
	return session.id;
}

func (session *Session) User() User {
	return session.user;
}

func (session *Session) SetUser(user User) {
	fmt.Printf("Session login: %s (%s)\n", user.DisplayName(), session.id)
	session.user = user
}
