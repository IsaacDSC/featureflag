package featureflag

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fields struct {
	host string
}

type args struct {
	key string
}

func TestFeatureFlagSDK_GetFeatureFlag(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//if r.URL.Path != "/fixedvalue" {
		//	t.Errorf("Expected to request '/fixedvalue', got: %s", r.URL.Path)
		//}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Expected Accept: application/json header, got: %s", r.Header.Get("Accept"))
		}

		mapper := map[string]struct {
			Active bool `json:"active"`
		}{
			"teste1": {Active: true},
			"teste2": {Active: false},
		}

		key := strings.Split(r.URL.Path, "/")[1]

		if _, ok := mapper[key]; !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		output, err := json.Marshal(mapper[key])
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		w.Write(output)
	}))

	defer server.Close()

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Should be retrieve active ff",
			fields: fields{
				host: server.URL,
			},
			args: args{
				key: "teste1",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "should be retrieve inactive ff",
			fields: fields{
				host: server.URL,
			},
			args: args{
				key: "teste2",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "should be retrieve error not found ff",
			fields: fields{
				host: server.URL,
			},
			args: args{
				key: "invalid",
			},
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ff := NewFeatureFlagSDK(tt.fields.host)

			got, err := ff.GetFeatureFlag(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFeatureFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetFeatureFlag() got = %v, want %v", got, tt.want)
			}
		})
	}
}
