package agent

import (
	"net/http"
	"net/url"
)

const (
	serviceHealth = "service"
	connectHealth = "connect"
	ingressHealth = "ingress"
)

func (s *HTTPHandlers) HealthChecksInState(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	return nil, nil

}

func (s *HTTPHandlers) HealthNodeChecks(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	return nil, nil
}

func (s *HTTPHandlers) HealthServiceChecks(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	return nil, nil
}

// HealthIngressServiceNodes should return "all the healthy ingress gateway instances
// that I can use to access this connect-enabled service without mTLS".
func (s *HTTPHandlers) HealthIngressServiceNodes(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	return nil, nil
}

// HealthConnectServiceNodes should return "all healthy connect-enabled
// endpoints (e.g. could be side car proxies or native instances) for this
// service so I can connect with mTLS".
func (s *HTTPHandlers) HealthConnectServiceNodes(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	return nil, nil
}

// HealthServiceNodes should return "all the healthy instances of this service
// registered so I can connect directly to them".
func (s *HTTPHandlers) HealthServiceNodes(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	return nil, nil
}

func (s *HTTPHandlers) healthServiceNodes(resp http.ResponseWriter, req *http.Request, healthType string) (interface{}, error) {
	return nil, nil
}

func getBoolQueryParam(params url.Values, key string) (bool, error) {
	return false, nil
}
