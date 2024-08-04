package http

import (
	"net"
	"net/http"
	"overkiz-adapter/internal/log"
	"strings"
)

func HostFilter(hosts ...string) func(next http.Handler) http.Handler {
	allowedHosts := make(map[string]struct{}, len(hosts))
	for _, host := range hosts {
		allowedHosts[strings.TrimSpace(strings.ToLower(host))] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			addresses := make([]string, 0)
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}
			addresses = append(addresses, host)
			lookupHosts, err := net.LookupAddr(host)
			if err == nil {
				addresses = append(addresses, lookupHosts...)
			}
			for _, address := range addresses {
				if _, ok := allowedHosts[strings.TrimSpace(strings.ToLower(address))]; ok {
					next.ServeHTTP(w, r)
					return
				}
			}
			log.Infof("Access forbidden for %v", addresses)
			w.WriteHeader(http.StatusForbidden)
			return
		}
		return http.HandlerFunc(fn)
	}
}
