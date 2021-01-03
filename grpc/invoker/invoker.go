package invoker

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

const (
	GrpcHeaderPath        = "Path"
	GrpcHeaderServiceName = "Service-Name"

	errMissingService       = "Requested service not found"
	errMissingServiceMethod = "Requested method not found"
)

type Invoker struct {
	Services map[string]ServiceInfo
}

type ServiceInfo struct {
	ServiceImpl interface{}
	Methods     map[string]grpc.MethodDesc
}

type GrpcRequest struct {
	Path        string
	ServiceName string
	Body        []byte
}

type GrpcResponse struct {
	Body []byte
}

func NewInvoker() *Invoker {
	return &Invoker{
		Services: map[string]ServiceInfo{},
	}
}

func (s *Invoker) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	info := ServiceInfo{
		ServiceImpl: impl,
		Methods:     map[string]grpc.MethodDesc{},
	}

	for _, v := range desc.Methods {
		path := fmt.Sprintf("/%s/%s", desc.ServiceName, v.MethodName)
		info.Methods[path] = v
	}

	s.Services[desc.ServiceName] = info
}

func (s *Invoker) Invoke(ctx context.Context, request GrpcRequest) (*GrpcResponse, error) {
	df := func(v interface{}) error {
		err := proto.Unmarshal(request.Body, v.(proto.Message))
		if err != nil {
			return err
		}

		return nil
	}

	serviceInfo, ok := s.Services[request.ServiceName]
	if ok == false {
		return nil, fmt.Errorf(errMissingService)
	}

	methodInfo, ok := serviceInfo.Methods[request.Path]
	if ok == false {
		return nil, fmt.Errorf(errMissingServiceMethod)
	}

	response, err := methodInfo.Handler(serviceInfo.ServiceImpl, ctx, df, nil)
	if err != nil {
		return nil, err
	}

	message, err := proto.Marshal(response.(proto.Message))
	if err != nil {
		return nil, err
	}

	return &GrpcResponse{
		Body: message,
	}, nil
}
