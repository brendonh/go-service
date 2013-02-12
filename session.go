package goservice

import (
	"sync"
	"fmt"

	"code.google.com/p/go-uuid/uuid"
)


func BasicSessionCreator(sessConn SessionConnection) Session {
	return &BasicSession {
		id: uuid.New(),
		user: nil,
		Mutex: new(sync.Mutex),
		connection: sessConn,
	}
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
	session.connection.Send(msg)
}