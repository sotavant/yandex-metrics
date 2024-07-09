package middleware

import (
	"context"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type IPChecker struct {
	trustedSubnet *net.IPNet
}

func NewIPChecker(trustedSubnet string) *IPChecker {
	if trustedSubnet == "" {
		return nil
	}

	_, ipNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		panic(err)
	}
	return &IPChecker{trustedSubnet: ipNet}
}

func (ip *IPChecker) CheckIP(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		reqIP := r.Header.Get("X-Real-IP")
		if reqIP == "" {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		IP := net.ParseIP(reqIP)
		if IP == nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		}

		if !ip.trustedSubnet.Contains(IP) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(f)
}

func (ip *IPChecker) CheckIPInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		ips := md["x-real-ip"]
		if len(ips) > 0 {
			IP := net.ParseIP(ips[0])
			if IP == nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid IP")
			}

			if !ip.trustedSubnet.Contains(IP) {
				return nil, status.Errorf(codes.Unauthenticated, "forbidden IP")
			}

			return handler(ctx, req)
		}

		return nil, status.Errorf(codes.Unauthenticated, "not found X-Real-IP")
	} else {
		return nil, status.Errorf(codes.Unauthenticated, "not found X-Real-IP")
	}
}
