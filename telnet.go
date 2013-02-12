package goservice

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strings"
	"encoding/json"
)

type telnetCommand func(*telnetConnection, []string)

type TelnetEndpoint struct {
	Address string
	context ServerContext
	listener net.Listener
	logPrefix string
	commands map[string]telnetCommand
}

type telnetConnection struct {
	endpoint *TelnetEndpoint
	conn net.Conn
	closed bool
	log func(string, ...interface{})
}

func NewTelnetEndpoint(address string, context ServerContext) Endpoint {
	var endpoint = &TelnetEndpoint{
		Address: address,
		context: context,
		logPrefix: "Telnet " + address,
		commands: make(map[string]telnetCommand),
	}
	endpoint.commands["quit"] = telnet_command_quit
	endpoint.commands["help"] = telnet_command_help
	return endpoint
}

func (endpoint *TelnetEndpoint) Start() bool {
	listener, err := net.Listen("tcp", endpoint.Address)
	if err != nil {
		endpoint.Log("Error listening on %s: %s", endpoint.Address, err)
		return false
	}

	endpoint.listener = listener

	go endpoint.Listen()

	endpoint.Log("Telnet endpoint started on %s", endpoint.Address)
	
	return true
}

func (endpoint *TelnetEndpoint) Listen() {
	for {
		conn, err := endpoint.listener.Accept()
		if err != nil {
			endpoint.Log("Error accepting: %s", err)
			continue
		}

		var prefix = fmt.Sprintf("[ %s ] ", conn.RemoteAddr())
		var log = func(fmt string, args... interface{}) {
			endpoint.Log(prefix + fmt, args...)
		}

		var tConn = &telnetConnection{
			endpoint: endpoint,
			conn: conn,
			closed: false,
			log: log,
		}
		go tConn.Loop()
	}
}

func (endpoint *TelnetEndpoint) Stop() bool {
	return true
}

func (endpoint *TelnetEndpoint) Log(fmt string, args... interface{}) {
	endpoint.context.LogPrefix(endpoint.logPrefix, fmt, args...)
}


func (tc *telnetConnection) Loop() {
	tc.log("Connection started")
	tc.WriteLinef("Type 'help' for help")

	var reader = textproto.NewReader(bufio.NewReader(tc.conn))
	for !tc.closed {
		tc.conn.Write([]byte("loge> "))

		line, err := reader.ReadLine()
		if err != nil {
			if err != io.EOF {
				tc.log("Connection error: %#v", err)
			}
			break
		}

		var tokens []string
		for _, tok := range strings.Split(string(line), " ") {
			if len(tok) > 0 {
				tokens = append(tokens, tok)
			}
		}

		if len(tokens) == 0 {
			continue
		}

		handler, ok := tc.endpoint.commands[tokens[0]]
		if ok {
			handler(tc, tokens[1:])
			continue
		}

		telnet_dispatch(tc, tokens)
	}
		
	tc.conn.Close()
	tc.log("Connection closed")
}

func (tc *telnetConnection) WriteLinef(format string, args... interface{}) {
	tc.conn.Write([]byte(fmt.Sprintf(format + "\n", args...)))
}

// ------------------------------------
// Commands
// ------------------------------------

func telnet_command_quit(tc *telnetConnection, args []string) {
	tc.WriteLinef("Bye!")
	tc.closed = true
}

func telnet_command_help(tc *telnetConnection, args []string) {
	var api = tc.endpoint.context.API()

	if len(args) == 0 {
		tc.WriteLinef("  help -- Show this help")
		tc.WriteLinef("  quit -- Close connection")

		var services = api.GetServices()
		if len(services) == 1 {
			tc.WriteLinef("")
			var service APIService
			for _, service = range api.GetServices() {
				break
			}
			telnet_list_service_commands(tc, service)
		} else {
			for name, service := range api.GetServices() {
				tc.WriteLinef("\nService '%s':\n", name)
				telnet_list_service_commands(tc, service)
			}
		}
		return
	}

	service, args := telnet_get_service(api, args)
	if service == nil || len(args) == 0 {
		tc.WriteLinef("Usage: help <service> <command>")
		return
	}

	var methodName = args[0]
	
	method, ok := service.GetMethods()[methodName]

	if !ok {
		tc.WriteLinef("Unknown method '%s'", methodName)
		return
	}

	tc.WriteLinef("%s:", methodName)
	if len(method.ArgSpec) == 0 {
		tc.WriteLinef("  (Takes no arguments)")
		return
	}

	for _, arg := range method.ArgSpec {
		tc.WriteLinef("  %s (%s)", arg.Name, stringArgType(arg.ArgType))
	}


}

func telnet_get_service(api API, args []string) (APIService, []string) {
	var services = api.GetServices()
	if len(services) == 1 {
		for _, service := range services {
			return service, args
		}
	}

	if len(args) == 0 {
		return nil, args
	}
	var serviceName string
	serviceName = args[0]
	service, ok := services[serviceName]
	if !ok {
		return nil, args
	}
	return service, args[1:]
}

func telnet_list_service_commands(tc *telnetConnection, service APIService) {
	for name, _ := range service.GetMethods() {
		tc.WriteLinef("  %s", name)
	}
}

func telnet_dispatch(tc *telnetConnection, args []string) {
	var api = tc.endpoint.context.API()

	var service APIService
	service, args = telnet_get_service(api, args)

	if len(args) == 0 {
		tc.WriteLinef("No command given")
		return
	}

	command, args := args[0], args[1:]

	method, ok := service.GetMethods()[command]

	if !ok {
		tc.WriteLinef("Unknown %s method '%s'", service.Name(), command)
		return
	}

	var mapArgs = make(APIData)
	for i, arg := range args {
		if i >= len(method.ArgSpec) {
			tc.WriteLinef("Too many arguments")
			return
		}
		mapArgs[method.ArgSpec[i].Name] = arg
	}

	ok, errors, funcArgs := Parse(method.ArgSpec, mapArgs)
	if !ok {
		tc.WriteLinef("Parse errors:")
		for _, line := range ListToStringSlice(errors) {
			tc.WriteLinef(" %s", line)
		}
		return
	}

	ok, response := method.Handler(funcArgs, nil, tc.endpoint.context)

	jsonResponse, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		tc.WriteLinef("Error encoding response: %s", err)
		return
	}

	tc.WriteLinef("%s", jsonResponse)
}
		
