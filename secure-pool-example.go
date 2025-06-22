package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

  "google.golang.org/api/option"
	"cloud.google.com/go/auth/grpctransport"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
  credentials "google.golang.org/genproto/googleapis/iam/credentials/v1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/alexflint/go-arg"
	"github.com/golang/protobuf/ptypes"
	"github.com/valkey-io/valkey-go"
)

type Rconf struct {
	Host string
	Port string
	Cert string
}

var (
	mu                 sync.RWMutex
	lastRefreshInstant = time.Time{}
	errLastSeen        = error(nil)
	token              = ""
)

var args struct {
	Project                  string        `help:"GCP ProjectID" default:"" arg:"--project, -p, env:GCP_PROJECT"`
	Instance                 string        `help:"Memorystore Instance name" default:"" arg:"--instance, -i, env:MEMORYSTORE_INSTANCE"`
	Lifetime                 time.Duration `help:"Lifetime of token" default:"1h" arg:"--lifetime, -l, env:LIFETIME"`
	RefreshDuration          time.Duration `help:"Refresh duration" default:"5m" arg:"--refresh-duration, -r, env:REFRESH_DURATION"`
	CheckTokenExpiryInterval time.Duration `help:"Check token expiry interval" default:"10s" arg:"--check-token-expiry-interval, -c, env:CHECK_TOKEN_EXPIRY_INTERVAL"`
}

func refreshTokenLoop() {
	if args.RefreshDuration > args.Lifetime {
		log.Fatal("Refresh should not happen after token is already expired.")
	}
	for {
		mu.RLock()
		lastRefreshTime := lastRefreshInstant
		mu.RUnlock()
		if time.Now().After(lastRefreshTime.Add(args.RefreshDuration)) {
			var err error
			retrievedToken, err := retrieveTokenFunc(valkey.AuthCredentialsContext{})
			mu.Lock()
			token = retrievedToken
			if err != nil {
				errLastSeen = err
			} else {
				lastRefreshInstant = time.Now()
			}
			mu.Unlock()
		}
		time.Sleep(args.CheckTokenExpiryInterval)
	}
}

func retrieveTokenFunc(yo valkey.AuthCredentialsContext) (valkey.AuthCredentials, error) {
  ctx := context.Background()
  conn, err := grpctransport.Dial(
    ctx,
    option.WithEndpoint("iamcredentials.googleapis.com:443"),
    option.WithScopes("https://www.googleapis.com/auth/cloud-platform"),
    )
  if err != nil {
    log.Printf("Failed to call API: %v", err)
    return valkey.AuthCredentials{}, err
  }
  client := credentials.NewIAMCredentialsClient(conn)
  req := credentials.GenerateAccessTokenRequest{
    Name: "projects/-/serviceAccounts/-",
    Scope: []string{"https://www.googleapis.com/auth/cloud-platform"},
    Lifetime: ptypes.DurationProto(args.Lifetime),
  }
  resp, err := client.GenerateAccessToken(ctx, req)
  if err != nil {
    log.Printf("Failed to call API: %v", err)
    return valkey.AuthCredentials{}, err
  }
	username := "default"
	return valkey.AuthCredentials{Username: username, Password: resp.AccessToken}, nil
}

func getSecret(projectID string, secretID string) (Rconf, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	cfg := Rconf{}
	if err != nil {
		return cfg, err
	}

	// Fetch CERT
	secret, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s-cert/versions/latest", projectID, secretID),
	})
	if err != nil {
		return cfg, err
	}
	cfg.Cert = string(secret.Payload.Data)

	// Fetch HOST
	secret, err = client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s-ip/versions/latest", projectID, secretID),
	})
	if err != nil {
		return cfg, err
	}
	cfg.Host = string(secret.Payload.Data)

	// Fetch PORT
	secret, err = client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s-port/versions/latest", projectID, secretID),
	})
	if err != nil {
		return cfg, err
	}
	cfg.Port = string(secret.Payload.Data)

	return cfg, nil
}

func valkeyConfig(cfg Rconf) valkey.ClientOption {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(cfg.Cert))
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	return valkey.ClientOption{
		InitAddress:       []string{addr, addr},
		AuthCredentialsFn: retrieveTokenFunc,
		TLSConfig: &tls.Config{
			RootCAs: caCertPool,
		},
		SendToReplicas: func(cmd valkey.Completed) bool {
			return (cmd.IsReadOnly() && rand.Intn(2) == 0)
		},
	}
}

func main() {
	arg.MustParse(&args)
	if args.Project == "" || args.Instance == "" {
		fmt.Println("Must specify --project and --instance")
		return
	}
	cfg, err := getSecret(args.Project, args.Instance)
	if err != nil {
		log.Fatalf("Failed to get secret: %v", err)
	}

	conf := valkeyConfig(cfg)

	ctx := context.Background()
	rdb, err := valkey.NewClient(conf)
	if err != nil {
		log.Fatalf("configuration failed: %v", err)
	}
	defer rdb.Close()
	err = rdb.Do(ctx, rdb.B().Set().Key("key").Value("val").Build()).Error()
	if err != nil {
		log.Fatalf("Failed to set key: %v", err)
	}
	val, err := rdb.Do(ctx, rdb.B().Get().Key("key").Build()).ToString()
	if err != nil {
		log.Fatalf("Failed to get key: %v", err)
	}
	fmt.Printf("Got value: %s\n", val)
}
