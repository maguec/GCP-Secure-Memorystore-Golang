package main

import (
	"context"
	"fmt"
	"log"

	"crypto/tls"
	"crypto/x509"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/alexflint/go-arg"
	"github.com/go-redis/redis/v9"
)

type Rconf struct {
	Host string
	Port string
	Cert string
	Auth string
}

var args struct {
	Project   string `help:"GCP ProjectID" default:"" arg:"--project, -p, env:GCP_PROJECT"`
	Instance  string `help:"Memorystore Instance name" default:"" arg:"--instance, -i, env:MEMORYSTORE_INSTANCE"`
}

func getSecret(projectID string, secretID string) (Rconf, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
  cfg := Rconf{}
	if err != nil {
		return cfg, err
	}

  // Fetch AUTH
	secret, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s-auth/versions/latest", projectID, secretID),
	})
	if err != nil {
		return cfg, err
	}
  cfg.Auth = string(secret.Payload.Data)

  // Fetch CERT
  secret, err = client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
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

func redisConfig(cfg Rconf) *redis.Options {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(cfg.Cert))
	return &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Auth,
		TLSConfig: &tls.Config{
			RootCAs: caCertPool,
		},
    // Connection Pooling Options
    // ALWAYS use connection pooling
    MinIdleConns: 1,      // Ensure that there is always at least 1 conn = set to number of workers in prod
    MaxIdleConns: 1,      // Don't have too many connections open set o to number of workers in prod + alph
    ConnMaxLifetime: 0,   // Stay open
    ConnMaxIdleTime: time.Minute, // Close connections after 1 minute of inactivity - change in prod
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

  conf := redisConfig(cfg)

	ctx := context.Background()
	rdb := redis.NewClient(conf)
	defer rdb.Close()
	err = rdb.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		log.Fatalf("Failed to set key: %v", err)
	}
	val, err := rdb.Get(ctx, "key").Result()
	if err != nil {
		log.Fatalf("Failed to get key: %v", err)
	}
	fmt.Printf("Got value: %s\n", val)
}
