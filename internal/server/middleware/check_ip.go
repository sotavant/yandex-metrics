package middleware

import (
	"net"
	"net/http"
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
