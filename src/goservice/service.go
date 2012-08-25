package goservice

import (
	"container/list"
)


type Service struct {
	name string
	Methods map[string]APIMethod
}

type ServiceCollection struct {
	Services map[string]APIService
}


func NewService(name string) *Service {
	return &Service{ 
		name: name,
	    Methods: make(map[string]APIMethod),
	}
}

func NewServiceCollection() *ServiceCollection {
	return &ServiceCollection{ 
		Services: make(map[string]APIService),
	}
}

func (service *Service) AddMethod(
	name string, 
	argSpec []APIArg,
	handler APIHandler) {
	service.Methods[name] = APIMethod{
		Name: name,
		ArgSpec: argSpec,
		Handler: handler, 
	}
}

func (service *Service) Name() string {
	return service.name
}

func (service *Service) FindMethod(methodName string) *APIMethod {
	method, ok := service.Methods[methodName]
	if !ok { 
		return nil
	}
	return &method
}


// ------------------------------------------
// Collections
// ------------------------------------------

func (collection *ServiceCollection) AddService(service APIService) {
	collection.Services[service.Name()] = service
}

var requestArgSpec = []APIArg {
	APIArg{Name: "service", ArgType: StringArg},
	APIArg{Name: "method", ArgType: StringArg},
	APIArg{Name: "data", ArgType: RawArg},
}

func (collection ServiceCollection) HandleRequest(request APIData, session *Session, context ServerContext) APIData {

	ok, resolutionErrors, args := Parse(requestArgSpec, request)
	if !ok {
		return ErrorResponse(ListToStringSlice(resolutionErrors))
	}

	return Response(collection.HandleCall(
		args["service"].(string), 
		args["method"].(string),
		args["data"].(APIData),
		session, context))

}

func (collection ServiceCollection) HandleCall(
	serviceName string, 
	methodName string,
	data APIData,
	session *Session,
	context ServerContext) (bool, []string, APIData) {


	service, ok := collection.Services[serviceName]
	if !ok {
		return false, []string{"No such service"}, nil
	}

	method := service.FindMethod(methodName)
	if method == nil {
		return false, []string{"No such method"}, nil
	}

	ok, errors, args := Parse(method.ArgSpec, data)
	if !ok {
		return false, ListToStringSlice(errors), nil
	}

	ok, response := method.Handler(args, session, context)
	if !ok {
		return false, nil, response
	}

	return true, nil, response
}

func ListToStringSlice(l *list.List) []string {
	var slice = make([]string, l.Len())
	var i = 0
	for el := l.Front(); el != nil; el = el.Next() {
		slice[i] = el.Value.(string)
		i++
	}
	return slice
}


func Response(ok bool, errors []string, response APIData) APIData {
	if ok { 
		return SuccessResponse(response)
	}

	if errors != nil {
		return ErrorResponse(errors)
	}

	return FailureResponse(response)
}

func ErrorResponse(errors []string) APIData {
	var response = make(APIData)
	response["success"] = false
	response["reason"] = "call error"
	response["errors"] = errors
	return response
}

func SuccessResponse(data APIData) APIData {
	var response = make(APIData)
	response["success"] = true
	response["data"] = data
	return response
}

func FailureResponse(errors APIData) APIData {
	var response = make(APIData)
	response["success"] = false
	response["reason"] = "failure"
	response["errors"] = errors
	return response
}