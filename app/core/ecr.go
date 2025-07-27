package core

import (
	"context"
	"embed"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/seriallink/datamaster/app/misc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// PublishDockerImages loads, tags, and pushes Docker images from embedded .tar artifacts to AWS ECR.
// If image names are provided, only those will be published. Otherwise, all .tar files in the artifacts
// directory are processed. The function requires Docker to be installed and available in the system PATH.
//
// Parameters:
//   - cfg: the AWS configuration used for ECR authentication and account resolution.
//   - artifacts: the embedded filesystem containing Docker .tar files.
//   - images: optional list of image base names (without extension) to publish.
//
// Returns:
//   - map[string]string: a mapping of image names to their full ECR URIs.
//   - error: an error if any step in the publishing process fails.
func PublishDockerImages(cfg aws.Config, artifacts embed.FS, images ...string) (map[string]string, error) {

	var (
		err       error
		files     []os.DirEntry
		identity  *sts.GetCallerIdentityOutput
		imageURIs = make(map[string]string)
	)

	files, err = artifacts.ReadDir(misc.ArtifactsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read artifacts directory: %w", err)
	}

	identity, err = GetCallerIdentity(context.TODO(), cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get caller identity: %w", err)
	}

	only := map[string]bool{}
	if len(images) > 0 {
		for _, name := range images {
			only[name+".tar"] = true
		}
	}

	for _, file := range files {

		var (
			name    = file.Name()
			content []byte
			tmpFile *os.File
		)

		if file.IsDir() || !strings.HasSuffix(name, ".tar") {
			continue
		}
		if len(only) > 0 && !only[name] {
			continue
		}

		imageName := strings.TrimSuffix(name, ".tar")
		tagName := misc.NameWithDefaultPrefix(imageName, '-')
		fullTag := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:latest", aws.ToString(identity.Account), cfg.Region, tagName)

		content, err = artifacts.ReadFile(path.Join(misc.ArtifactsPath, name))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", name, err)
		}

		tmpFile, err = os.CreateTemp("", "*.tar")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp file: %w", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err = tmpFile.Write(content); err != nil {
			return nil, fmt.Errorf("failed to write temp file: %w", err)
		}
		if err = tmpFile.Close(); err != nil {
			return nil, fmt.Errorf("failed to close temp file: %w", err)
		}

		if _, err = exec.LookPath("docker"); err != nil {
			return nil, fmt.Errorf("docker not found in PATH: %w", err)
		}

		fmt.Println(misc.Blue("Loading Docker image %s...", imageName))
		if err = exec.Command("docker", "load", "-i", tmpFile.Name()).Run(); err != nil {
			return nil, fmt.Errorf("failed to load docker image %s: %w", imageName, err)
		}

		fmt.Println(misc.Blue("Tagging image %s as %s...", imageName, fullTag))
		if err = exec.Command("docker", "tag", imageName+":latest", fullTag).Run(); err != nil {
			return nil, fmt.Errorf("failed to tag docker image %s: %w", imageName, err)
		}

		if len(imageURIs) == 0 {
			if err := LoginToECRWithSDK(cfg); err != nil {
				return nil, fmt.Errorf("failed to authenticate with ECR via SDK: %w", err)
			}
		}

		fmt.Println(misc.Blue("Pushing image %s...", fullTag))
		push := exec.Command("docker", "push", fullTag)
		push.Stdout = os.Stdout
		push.Stderr = os.Stderr
		if err = push.Run(); err != nil {
			return nil, fmt.Errorf("failed to push docker image %s: %w", imageName, err)
		}

		imageURIs[imageName] = fullTag
	}

	if len(imageURIs) == 0 {
		return nil, fmt.Errorf("no Docker images were published")
	}

	return imageURIs, nil

}

// LoginToECRWithSDK authenticates the Docker CLI with AWS ECR using credentials retrieved via the AWS SDK.
// It fetches the ECR authorization token, decodes it, and performs a `docker login` using `--password-stdin`.
//
// Parameters:
//   - cfg: the AWS configuration used to create the ECR client.
//
// Returns:
//   - error: an error if the token retrieval, decoding, or Docker login fails.
func LoginToECRWithSDK(cfg aws.Config) error {

	client := ecr.NewFromConfig(cfg)

	resp, err := client.GetAuthorizationToken(context.TODO(), &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return fmt.Errorf("failed to get authorization token: %w", err)
	}

	if len(resp.AuthorizationData) == 0 {
		return fmt.Errorf("no authorization data received")
	}

	auth := resp.AuthorizationData[0]
	token := aws.ToString(auth.AuthorizationToken)
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return fmt.Errorf("failed to decode token: %w", err)
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("unexpected token format")
	}
	username, password := parts[0], parts[1]
	endpoint := aws.ToString(auth.ProxyEndpoint)

	fmt.Println(misc.Blue("Logging in to ECR via SDK..."))

	cmd := exec.Command("docker", "login",
		"--username", username,
		"--password-stdin", endpoint)
	cmd.Stdin = strings.NewReader(password)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()

}
