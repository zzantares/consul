package dataplane

import (
	"context"

	"github.com/hashicorp/consul/proto-public/pbdataplane"
)

func (d *Server) SupportedDataplaneFeatures(ctx context.Context, req *pbdataplane.SupportedDataplaneFeaturesRequest) (*pbdataplane.SupportedDataplaneFeaturesResponse, error) {
	// TODO: Do token authorization once the blocking task is complete - https://app.asana.com/0/1200944825835228/1201971577284008/f
	// authz, err := d.Backend.ResolveTokenAndDefaultMeta(token)
	// if err != nil {
	// 	return err
	// }
	// if err := authz.ToAllowAuthorizer().ServiceWriteAnyAllowed(&authzContext); err != nil {
	// 	return err
	// }

	supportedFeatures := []*pbdataplane.DataplaneFeatureSupport{
		{
			FeatureName: pbdataplane.DataplaneFeatures_WATCH_SERVERS,
			Supported:   true,
		},
		{
			FeatureName: pbdataplane.DataplaneFeatures_EDGE_CERTIFICATE_MANAGEMENT,
			Supported:   true,
		},
		{
			FeatureName: pbdataplane.DataplaneFeatures_ENVOY_BOOTSTRAP_CONFIGURATION,
			Supported:   true,
		},
	}
	return &pbdataplane.SupportedDataplaneFeaturesResponse{SupportedDataplaneFeatures: supportedFeatures}, nil
}
