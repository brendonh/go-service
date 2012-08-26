package goservice

import (
	"fmt"
	"io"
	"reflect"
	"bytes"
	"net"
	"net/http"

	"github.com/ugorji/go-msgpack"
	"code.google.com/p/go.net/websocket"
)

type MessageHandler func(*WebsocketEndpoint, []byte, Session, *websocket.Conn)

type WebsocketEndpoint struct {
	Address string
	listener net.Listener
	context ServerContext
	Handler MessageHandler
}


func NewWebsocketEndpoint(address string, context ServerContext) *WebsocketEndpoint {
	return &WebsocketEndpoint{
		Address: address,
		context: context,
		Handler: DefaultMessageHandler,
	}
}

func DefaultMessageHandler(endpoint *WebsocketEndpoint, buf []byte, session Session, conn *websocket.Conn) {
	endpoint.HandleAPI(buf, session, conn)
}

func (endpoint *WebsocketEndpoint) Start() bool {
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

	var handler = func(ws *websocket.Conn) {
		endpoint.Handle(ws)
	}

	mux.Handle("/", websocket.Handler(handler))
	go http.Serve(listener, mux)

	return true
}


func (endpoint *WebsocketEndpoint) Stop() bool {
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


func (endpoint *WebsocketEndpoint) Handle(ws *websocket.Conn) {
	ws.PayloadType = websocket.BinaryFrame

	var buf = make([]byte, 1024 * 64)
	var session Session = NewBasicSession()

	fmt.Printf("New session: %s\n", session.ID())

	for {

		msgLength, err := ws.Read(buf)
		
		if err != nil {
			if err != io.EOF {
				fmt.Printf("WS error: %#v\n", err)
			}
			fmt.Printf("Session closed: %s\n", session.ID())
			break
		}

		if msgLength == 0 {
			continue
		}

		var msgBuf = make([]byte, msgLength)
		copy(msgBuf, buf)

		endpoint.Handler(endpoint, msgBuf, session, ws)
	}
}


func (endpoint *WebsocketEndpoint) HandleAPI(buf []byte, session Session, ws *websocket.Conn) {
	var data APIData
	var resolver = msgpack.DefaultDecoderContainerResolver
	resolver.MapType = reflect.TypeOf(make(APIData))

	var dec = msgpack.NewDecoder(bytes.NewReader(buf), &resolver)
	
	var err = dec.Decode(&data)

	if err != nil {
		fmt.Printf("Decode err: %v\n", err)
		return
	}

	var response = endpoint.context.API().HandleRequest(data, session, endpoint.context)

	if id, ok := data["id"]; ok {
		response["id"] = id
	}

	w := bytes.NewBufferString("")
	w.WriteByte('a')
	enc := msgpack.NewEncoder(w)
	err = enc.Encode(response)

	if err != nil {
		fmt.Printf("Encode err: %#v\n", err)
		return
	}

	ws.Write(w.Bytes())
}