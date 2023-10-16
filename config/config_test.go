package config

import (
	"os"
	"reflect"
	"testing"
)

func Test_replaceEnvVars(t *testing.T) {
	type args struct {
		rawConfig []byte
	}
	tests := []struct {
		name     string
		args     args
		envName  string
		envValue string
		want     []byte
	}{
		{
			name: "test",
			args: args{
				rawConfig: []byte(`{
{
  "apiGateway": {
    "name": "MyAPIGateway",
    "version": "1.0.0"
  },
  "vhosts": {
    "*": {
      "endpoints": [
        {
          "path": "/ws",
          "backend": {
            "url": "ws://127.0.0.1:8080/ws/$TEST"
          },
          "websocket": {
            "readBufferSize": 1024,
            "writeBufferSize": 1024,
            "allowedOrigins": ["*"]
          }
        }
      ]
    }
  }
}
}`),
			},
			envName:  "TEST",
			envValue: "test",
			want: []byte(`{
{
  "apiGateway": {
    "name": "MyAPIGateway",
    "version": "1.0.0"
  },
  "vhosts": {
    "*": {
      "endpoints": [
        {
          "path": "/ws",
          "backend": {
            "url": "ws://127.0.0.1:8080/ws/test"
          },
          "websocket": {
            "readBufferSize": 1024,
            "writeBufferSize": 1024,
            "allowedOrigins": ["*"]
          }
        }
      ]
    }
  }
}
}`),
		},
		{
			name: "path",
			args: args{
				rawConfig: []byte(`{
{
  "apiGateway": {
    "name": "MyAPIGateway",
    "version": "1.0.0"
  },
  "vhosts": {
    "*": {
      "endpoints": [
        {
          "path": "/ws",
          "backend": {
            "url": "ws://127.0.0.1:8080/ws"
          },
          "websocket": {
            "readBufferSize": 1024,
            "writeBufferSize": 1024,
            "allowedOrigins": ["*"]
          }
        }
      ]
    }
  }
}
}`),
			},
			want: []byte(`{
{
  "apiGateway": {
    "name": "MyAPIGateway",
    "version": "1.0.0"
  },
  "vhosts": {
    "*": {
      "endpoints": [
        {
          "path": "/ws",
          "backend": {
            "url": "ws://127.0.0.1:8080/ws"
          },
          "websocket": {
            "readBufferSize": 1024,
            "writeBufferSize": 1024,
            "allowedOrigins": ["*"]
          }
        }
      ]
    }
  }
}
}`),
			envName:  "path",
			envValue: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv(tt.envName, tt.envValue)
			if got := replaceEnvVars(tt.args.rawConfig); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("replaceEnvVars() = %v, wantAny %v", got, tt.want)
			}
		})
	}
}

func Test_checkVhostsWithTLS(t *testing.T) {
	type args struct {
		config *Config
	}
	tests := []struct {
		name    string
		args    args
		wantAny bool
		wantAll bool
	}{
		{
			name: "test",
			args: args{
				config: &Config{Vhosts: map[string]Vhost{
					"www.sample.org": {TLS: &TLSConfig{}},
				}},
			},
			wantAny: true,
			wantAll: true,
		},
		{
			name: "test",
			args: args{
				config: &Config{Vhosts: map[string]Vhost{
					"www.sample.org":  {TLS: &TLSConfig{}},
					"www1.sample.org": {},
				}},
			},
			wantAny: true,
			wantAll: false,
		},
		{
			name: "test",
			args: args{
				config: &Config{Vhosts: map[string]Vhost{
					"www.sample.org":  {},
					"www1.sample.org": {},
				}},
			},
			wantAny: false,
			wantAll: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := checkVhostsWithTLS(tt.args.config)
			if got != tt.wantAny {
				t.Errorf("checkVhostsWithTLS() got = %v, wantAny %v", got, tt.wantAny)
			}
			if got1 != tt.wantAll {
				t.Errorf("checkVhostsWithTLS() got1 = %v, wantAny %v", got1, tt.wantAll)
			}
		})
	}
}
