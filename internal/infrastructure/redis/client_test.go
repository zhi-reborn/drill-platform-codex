package redis

import (
	"crypto/tls"
	"reflect"
	"testing"

	goredis "github.com/redis/go-redis/v9"
)

func TestNewUniversalClientUsesClusterWhenConfigured(t *testing.T) {
	client := newUniversalClient(Config{
		Addr:         "standalone:6379",
		ClusterAddrs: "redis-a:6379, redis-b:6379",
		Username:     "drill",
		Password:     "secret",
		TLS:          true,
		PoolSize:     12,
	})
	defer client.Close()

	cluster, ok := client.(*goredis.ClusterClient)
	if !ok {
		t.Fatalf("client type = %T, want *redis.ClusterClient", client)
	}

	opts := cluster.Options()
	if !opts.ContextTimeoutEnabled {
		t.Fatal("cluster must honor context deadlines for presence requests")
	}
	if !reflect.DeepEqual(opts.Addrs, []string{"redis-a:6379", "redis-b:6379"}) {
		t.Fatalf("cluster addrs = %#v, want redis-a/redis-b", opts.Addrs)
	}
	if opts.Username != "drill" || opts.Password != "secret" {
		t.Fatalf("cluster auth = %q/%q, want drill/secret", opts.Username, opts.Password)
	}
	if opts.TLSConfig == nil {
		t.Fatal("cluster TLSConfig = nil, want configured TLS")
	}
	if opts.PoolSize != 12 {
		t.Fatalf("cluster PoolSize = %d, want 12", opts.PoolSize)
	}
}

func TestNewUniversalClientUsesSentinelWhenConfigured(t *testing.T) {
	client := newUniversalClient(Config{
		Addr:           "sentinel-a:26379,sentinel-b:26379",
		SentinelMaster: "drill-master",
		Username:       "drill",
		Password:       "secret",
		DB:             2,
		TLS:            true,
	})
	defer client.Close()

	standalone, ok := client.(*goredis.Client)
	if !ok {
		t.Fatalf("client type = %T, want *redis.Client", client)
	}

	opts := standalone.Options()
	if !opts.ContextTimeoutEnabled {
		t.Fatal("sentinel must honor context deadlines for presence requests")
	}
	if opts.Addr != "FailoverClient" {
		t.Fatalf("sentinel client addr = %q, want FailoverClient", opts.Addr)
	}
	if opts.Username != "drill" || opts.Password != "secret" {
		t.Fatalf("sentinel auth = %q/%q, want drill/secret", opts.Username, opts.Password)
	}
	if opts.DB != 2 {
		t.Fatalf("sentinel DB = %d, want 2", opts.DB)
	}
	if opts.TLSConfig == nil {
		t.Fatal("sentinel TLSConfig = nil, want configured TLS")
	}
}

func TestClientModeReportsConfiguredTopology(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "cluster",
			cfg:  Config{ClusterAddrs: "redis-a:6379,redis-b:6379"},
			want: "cluster",
		},
		{
			name: "sentinel",
			cfg:  Config{Addr: "sentinel-a:26379", SentinelMaster: "drill-master"},
			want: "sentinel",
		},
		{
			name: "standalone",
			cfg:  Config{Addr: "redis:6379"},
			want: "standalone",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := &Client{rc: newUniversalClient(tc.cfg)}
			defer client.Close()

			if got := client.Mode(); got != tc.want {
				t.Fatalf("Mode() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRedisTLSConfig(t *testing.T) {
	standalone := newUniversalClient(Config{Addr: "localhost:6379"})
	defer standalone.Close()
	if !standalone.(*goredis.Client).Options().ContextTimeoutEnabled {
		t.Fatal("standalone must honor context deadlines for presence requests")
	}
	if redisTLSConfig(false) != nil {
		t.Fatal("redisTLSConfig(false) returned non-nil")
	}
	cfg := redisTLSConfig(true)
	if cfg == nil {
		t.Fatal("redisTLSConfig(true) returned nil")
	}
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Fatalf("MinVersion = %d, want TLS 1.2", cfg.MinVersion)
	}
}
