package xds

import (
	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/hashicorp/go-version"
)

var (
	// minSupportedVersion is the oldest mainline version we support. This should always be
	// the zero'th point release of the last element of proxysupport.EnvoyVersions.
	minSupportedVersion = version.Must(version.NewVersion("1.13.0"))

	specificUnsupportedVersions = []unsupportedVersion{
		{
			Version:   version.Must(version.NewVersion("1.13.0")),
			UpgradeTo: "1.13.1+",
			Why:       "does not support RBAC rules using url_path",
		},
	}
)

type unsupportedVersion struct {
	Version   *version.Version
	UpgradeTo string
	Why       string
}

type supportedProxyFeatures struct {
	// add version dependent feature flags here
}

func determineSupportedProxyFeatures(node *envoycore.Node) (supportedProxyFeatures, error) {
	return supportedProxyFeatures{}, nil
}

func determineSupportedProxyFeaturesFromString(vs string) (supportedProxyFeatures, error) {
	return supportedProxyFeatures{}, nil
}

func determineSupportedProxyFeaturesFromVersion(version *version.Version) (supportedProxyFeatures, error) {
	return supportedProxyFeatures{}, nil
}

// example: 1580db37e9a97c37e410bad0e1507ae1a0fd9e77/1.12.4/Clean/RELEASE/BoringSSL

func determineEnvoyVersionFromNode(node *envoycore.Node) *version.Version {
	return nil
}
