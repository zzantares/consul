package connect

import (
	"fmt"

	"github.com/hashicorp/consul/agent/structs"
)

type UpstreamSNIOpts struct {
	DestinationType    string
	Service            string
	Subset             string
	Namespace          string
	LocalDatacenter    string
	UpstreamDatacenter string
	TrustDomain        string
}

func UpstreamSNI(opts UpstreamSNIOpts) string {
	dc := opts.LocalDatacenter
	if opts.UpstreamDatacenter != "" {
		dc = opts.UpstreamDatacenter
	}

	if opts.DestinationType == structs.UpstreamDestTypePreparedQuery {
		return QuerySNI(opts.Service, dc, opts.TrustDomain)
	}
	return ServiceSNI(opts.Service, opts.Subset, opts.Namespace, dc, opts.TrustDomain)
}

func DatacenterSNI(dc string, trustDomain string) string {
	return fmt.Sprintf("%s.internal.%s", dc, trustDomain)
}

func ServiceSNI(service string, subset string, namespace string, datacenter string, trustDomain string) string {
	if namespace == "" {
		namespace = "default"
	}

	if subset == "" {
		return fmt.Sprintf("%s.%s.%s.internal.%s", service, namespace, datacenter, trustDomain)
	} else {
		return fmt.Sprintf("%s.%s.%s.%s.internal.%s", subset, service, namespace, datacenter, trustDomain)
	}
}

func QuerySNI(service string, datacenter string, trustDomain string) string {
	return fmt.Sprintf("%s.default.%s.query.%s", service, datacenter, trustDomain)
}

func TargetSNI(target *structs.DiscoveryTarget, trustDomain string) string {
	return ServiceSNI(target.Service, target.ServiceSubset, target.Namespace, target.Datacenter, trustDomain)
}
