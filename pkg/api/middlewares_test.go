package api

import "testing"

func Test_extractNamespaceAndTopic(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		wantNamespace string
		wantTopic     string
	}{
		{
			name:          "case /namespaces",
			path:          "api/v2/namespaces",
			wantNamespace: "",
			wantTopic:     "namespaces",
		},
		{
			name:          "case /namespaces",
			path:          "/api/v2/namespaces",
			wantNamespace: "",
			wantTopic:     "namespaces",
		},
		{
			name:          "case /namespaces",
			path:          "/api/v2/namespaces/",
			wantNamespace: "",
			wantTopic:     "namespaces",
		},
		{
			name:          "case /namespaces",
			path:          "//api/v2//namespaces//",
			wantNamespace: "",
			wantTopic:     "namespaces",
		},
		{
			name:          "case /namespaces",
			path:          "//api/v2//namespaces//",
			wantNamespace: "",
			wantTopic:     "namespaces",
		},

		{
			name:          "without /namespaces",
			path:          "api/v2//",
			wantNamespace: "",
			wantTopic:     "",
		},

		{
			name:          "without /namespaces",
			path:          "api/v2",
			wantNamespace: "",
			wantTopic:     "",
		},

		{
			name:          "case /namespaces/123",
			path:          "api/v2/namespaces/123",
			wantNamespace: "123",
			wantTopic:     "namespaces",
		},

		{
			name:          "case /namespaces/123",
			path:          "/api/v2/namespaces/123//",
			wantNamespace: "123",
			wantTopic:     "namespaces",
		},

		{
			name:          "case /namespaces/123",
			path:          "/api/v2/namespaces/123//",
			wantNamespace: "123",
			wantTopic:     "namespaces",
		},

		{
			name:          "case /logging",
			path:          "/api/v2/logging",
			wantNamespace: "",
			wantTopic:     "logging",
		},
		{
			name:          "case /logging",
			path:          "/api/v2/logging/",
			wantNamespace: "",
			wantTopic:     "logging",
		},

		{
			name:          "case /roles",
			path:          "/api/v2/roles",
			wantNamespace: "",
			wantTopic:     "roles",
		},
		{
			name:          "case /roles",
			path:          "/api/v2/roles/",
			wantNamespace: "",
			wantTopic:     "roles",
		},

		{
			name:          "case /namespaces/123/foo",
			path:          "/api/v2/namespaces/123/foo",
			wantNamespace: "123",
			wantTopic:     "foo",
		},

		{
			name:          "case /namespaces/123/foo",
			path:          "/api/v2/namespaces/123/foo/bar",
			wantNamespace: "123",
			wantTopic:     "foo",
		},
		{
			name:          "case /namespaces/p2/t2/something",
			path:          "/api/v2/namespaces/p2/t2/something",
			wantNamespace: "p2",
			wantTopic:     "t2",
		},
		{
			name:          "case /namespaces/api",
			path:          "/api/v2/namespaces/api/t2/something",
			wantNamespace: "api",
			wantTopic:     "t2",
		},
		{
			name:          "case /namespaces/api",
			path:          "/api/v2/namespaces/v1/t2/something",
			wantNamespace: "v1",
			wantTopic:     "t2",
		},
		{
			name:          "case /namespaces/api",
			path:          "/api/v2/namespaces/v2/t2/something",
			wantNamespace: "v2",
			wantTopic:     "t2",
		},
		{
			name:          "case /namespaces/namespaces",
			path:          "/api/v2/namespaces/namespaces/t2/something",
			wantNamespace: "namespaces",
			wantTopic:     "t2",
		},
		{
			name:          "case /namespaces/namespace",
			path:          "/api/v2/namespaces/namespace/t2/something",
			wantNamespace: "namespace",
			wantTopic:     "t2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNamespace, gotTopic := extractNamespaceAndTopic(tt.path)
			if gotNamespace != tt.wantNamespace {
				t.Errorf("extractNamespaceAndTopic() gotNamespace = %v, wantNamespace %v", gotNamespace, tt.wantNamespace)
			}
			if gotTopic != tt.wantTopic {
				t.Errorf("extractNamespaceAndTopic() gotTopic = %v, wantTopic %v", gotTopic, tt.wantTopic)
			}
		})
	}
}
