package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_createAllowedOriginFunc(t *testing.T) {
	type args struct {
		allowedDomains []string
		origin         string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "HappyPath - localhost with HTTP",
			args: args{
				allowedDomains: []string{"my-domain.com"},
				origin:         "http://localhost",
			},
			want: true,
		},
		{
			name: "HappyPath - localhost with HTTPS",
			args: args{
				allowedDomains: []string{"my-domain.com"},
				origin:         "https://localhost",
			},
			want: true,
		},
		{
			name: "HappyPath - localhost with port",
			args: args{
				allowedDomains: []string{"my-domain.com"},
				origin:         "http://localhost:8080",
			},
			want: true,
		},
		{
			name: "HappyPath - 0.0.0.0",
			args: args{
				allowedDomains: []string{"my-domain.com"},
				origin:         "http://0.0.0.0:8080",
			},
			want: true,
		},
		{
			name: "HappyPath - origin with port is in allowed domains",
			args: args{
				allowedDomains: []string{"my-domain.com"},
				origin:         "https://my-domain.com:8080",
			},
			want: true,
		},
		{
			name: "HappyPath - origin with sub-domain",
			args: args{
				allowedDomains: []string{"my-domain.com"},
				origin:         "https://www.my-domain.com",
			},
			want: true,
		},
		{
			name: "HappyPath - origin without port is in allowed domains",
			args: args{
				allowedDomains: []string{"my-domain.com"},
				origin:         "https://my-domain.com",
			},
			want: true,
		},
		{
			name: "Origin is not in allowed domains",
			args: args{
				allowedDomains: []string{"my-domain.com"},
				origin:         "https://not-my-domain.com",
			},
			want: false,
		},
		{
			name: "Origin contains allowed domains but not within the URL host",
			args: args{
				allowedDomains: []string{"my-domain.com"},
				origin:         "https://not-my-domain.com/my-domain.com",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowedFunc := createAllowedOriginFunc(tt.args.allowedDomains)
			allowed := allowedFunc(tt.args.origin)
			assert.Equal(t, tt.want, allowed)
		})
	}
}
