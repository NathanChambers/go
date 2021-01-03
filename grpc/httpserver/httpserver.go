package apigatewayserver

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/NathanChambers/go/grpc/invoker"
)

const (
	errMissingPathHeader        = "HTTP requests must be made using a 'Path' HTTP header."
	errMissingServiceNameHeader = "HTTP requests must be made using a 'Service-Name' HTTP header."
)

type HttpServer struct {
	*invoker.Invoker
}

func NewHttpServer() *HttpServer {
	return &HttpServer{
		Invoker: invoker.NewInvoker(),
	}
}

func (s *HttpServer) Handler(w http.ResponseWriter, request *http.Request) {
	path := request.Header.Get(invoker.GrpcHeaderPath)
	if path == "" {
		s.errorResponse(w, http.StatusBadRequest, fmt.Errorf(errMissingPathHeader))
		return
	}

	serviceName := request.Header.Get(invoker.GrpcHeaderServiceName)
	if serviceName == "" {
		s.errorResponse(w, http.StatusBadRequest, fmt.Errorf(errMissingServiceNameHeader))
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		s.errorResponse(w, http.StatusBadRequest, err)
		return
	}

	response, err := s.Invoker.Invoke(request.Context(), invoker.GrpcRequest{
		Path:        path,
		ServiceName: serviceName,
		Body:        body,
	})

	if err != nil {
		s.errorResponse(w, http.StatusBadRequest, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response.Body)
}

func (s *HttpServer) errorResponse(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	w.Write([]byte(err.Error()))
}
