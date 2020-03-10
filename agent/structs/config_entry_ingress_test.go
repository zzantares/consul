package structs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIngressConfigEntry_Validate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		entry     IngressGatewayConfigEntry
		expectErr string
	}{
		{
			name: "port conflict",
			entry: IngressGatewayConfigEntry{
				Kind: "ingress-gateway",
				Name: "ingress-web",
				Listeners: []IngressListener{
					IngressListener{
						Port:     1111,
						Protocol: "tcp",
						Services: []IngressService{
							IngressService{
								Name: "mysql",
							},
						},
					},
					IngressListener{
						Port:     1111,
						Protocol: "tcp",
						Services: []IngressService{
							IngressService{
								Name: "postgres",
							},
						},
					},
				},
			},
			expectErr: "port 1111 declared on two listeners",
		},
	}

	for _, tc := range cases {
		err := tc.entry.Validate()
		if tc.expectErr != "" {
			require.Error(t, err)
			requireContainsLower(t, err.Error(), tc.expectErr)
		} else {
			require.NoError(t, err)
		}
	}
}
