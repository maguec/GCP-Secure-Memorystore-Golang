package main

import (
	"context"
	"fmt"
	"log"

	"crypto/tls"
	"crypto/x509"

	memorystore "cloud.google.com/go/redis/apiv1"
	redispb "cloud.google.com/go/redis/apiv1/redispb"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/alexflint/go-arg"
	"github.com/go-redis/redis/v9"
)

var args struct {
	Project   string `help:"GCP ProjectID" default:"" arg:"--project, -p, env:GCP_PROJECT"`
	Instance  string `help:"Memorystore Instance name" default:"" arg:"--instance, -i, env:MEMORYSTORE_INSTANCE"`
	Loacation string `help:"Memorystore Instance location" default:"" arg:"--location, -l, env:MEMORYSTORE_LOCATION"`
}

func getSecret(projectID string, secretID string) (string, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", err
	}
	secret, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, secretID),
	})
	if err != nil {
		return "", err
	}
	return string(secret.Payload.Data), nil
}

func getInstance(projectID string, location string, instanceID string) (*redispb.Instance, error) {
	ctx := context.Background()
	client, err := memorystore.NewCloudRedisClient(ctx)
	if err != nil {
		return nil, err
	}
	instance, err := client.GetInstance(ctx, &redispb.GetInstanceRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/instances/%s", projectID, location, instanceID),
	})
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func redisConfig(instance *redispb.Instance, password string) *redis.Options {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(instance.ServerCaCerts[0].Cert))
	return &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", instance.Host, instance.Port),
		Password: password,
		TLSConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}
}

func main() {
	arg.MustParse(&args)
	if args.Project == "" || args.Instance == "" || args.Loacation == "" {
		fmt.Println("Must specify --project, --location and --instance")
		return
	}
	password, err := getSecret(args.Project, args.Instance)
	if err != nil {
		log.Fatalf("Failed to get secret: %v", err)
	}
	instance, err := getInstance(args.Project, args.Loacation, args.Instance)
	if err != nil {
		log.Fatalf("Failed to get instance: %v", err)
	}
  conf := redisConfig(instance, password)

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
