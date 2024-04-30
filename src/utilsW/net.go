package utilsW

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strings"
)

func BindDomain(bindMap map[string]string) *http.Client {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				for domainAddr, ip := range bindMap {
					if strings.Contains(addr, domainAddr) {
						return net.Dial(network, ip)
					}
				}
				return net.Dial(network, addr)
			},
		},
	}
	return &client
}
