package main

import (
  "context"
  "fmt"
  "github.com/alexflint/go-arg"
  secretmanager "cloud.google.com/go/secretmanager/apiv1"
  "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

var args struct {
        Project    string `help:"GCP ProjectID" default:"" arg:"--project, -p, env:GCP_PROJECT"`
        Instance string `help:"Memorystore Instance name" default:"" arg:"--instance, -i, env:MEMORYSTORE_INSTANCE"`
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

func main() {
  arg.MustParse(&args)
  if args.Project == "" || args.Instance == "" {
    fmt.Println("Must specify --project and --instance")
    return
  }
  fmt.Println(getSecret(args.Project, args.Instance))
}
