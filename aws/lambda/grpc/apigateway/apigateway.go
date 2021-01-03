package apigateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/NathanChambers/go/grpc/invoker"
	"github.com/aws/aws-lambda-go/events"
)

const (
	errMissingPathHeader        = "HTTP requests must be made using a 'Path' HTTP header."
	errMissingServiceNameHeader = "HTTP requests must be made using a 'Service-Name' HTTP header."
)

type ApiGatewayServer struct {
	*invoker.Invoker
}

func NewServer() *ApiGatewayServer {
	return &ApiGatewayServer{
		Invoker: invoker.NewInvoker(),
	}
}

func (s *ApiGatewayServer) Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	path, ok := request.Headers[invoker.GrpcHeaderPath]
	if ok == false {
		return s.serviceError(http.StatusBadRequest, fmt.Errorf(errMissingPathHeader))
	}

	serviceName, ok := request.Headers[invoker.GrpcHeaderServiceName]
	if ok == false {
		return s.serviceError(http.StatusBadRequest, fmt.Errorf(errMissingServiceNameHeader))
	}

	response, err := s.Invoker.Invoke(ctx, invoker.GrpcRequest{
		Path:        path,
		ServiceName: serviceName,
		Body:        []byte(request.Body),
	})

	if err != nil {
		return s.serviceError(http.StatusInternalServerError, err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(response.Body),
	}, nil
}

func (s *ApiGatewayServer) serviceError(status int, err error) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       err.Error(),
	}, err
}
