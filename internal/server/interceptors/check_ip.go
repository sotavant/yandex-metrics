package interceptors

import (
	"context"

	"google.golang.org/grpc"
)

func CheckIPInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

}
