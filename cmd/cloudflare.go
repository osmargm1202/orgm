package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/viper"
)

// r2HTTPGet downloads a file from a public R2 URL using token if provided
func r2HTTPGet(ctx context.Context, url string, token string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil { return nil, err }
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

// r2S3Client builds an S3 client for R2 using Viper credentials for a specific bucket
func r2S3Client(ctx context.Context, bucketKey string) (*s3.Client, error) {
	endpoint := viper.GetString(fmt.Sprintf("cloudflare.r2_%s.CLOUDFLARE_R2_ENDPOINT_URL", bucketKey))
	id := viper.GetString(fmt.Sprintf("cloudflare.r2_%s.CLOUDFLARE_R2_ACCESS_KEY_ID", bucketKey))
	secret := viper.GetString(fmt.Sprintf("cloudflare.r2_%s.CLOUDFLARE_R2_SECRET_ACCESS_KEY", bucketKey))

	// Fallback to old format if new format is not found
	if endpoint == "" {
		endpoint = viper.GetString("cloudflare.bucket.orgm-privado.s3_endpoint")
	}
	if id == "" {
		id = viper.GetString("cloudflare.bucket.orgm-privado.s3_id")
	}
	if secret == "" {
		secret = viper.GetString("cloudflare.bucket.orgm-privado.s3_secret")
	}

	if endpoint == "" || id == "" || secret == "" {
		return nil, fmt.Errorf("missing R2 S3 credentials or endpoint in viper for bucket %s", bucketKey)
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(id, secret, "")),
		awsconfig.WithRegion("auto"),
	)
	if err != nil { return nil, err }

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})
	return client, nil
}

// resolveConfigDir returns viper.config_path or ~/.config/orgm
func resolveConfigDir() (string, error) {
	if base := viper.GetString("config_path"); base != "" {
		return base, nil
	}
	home, err := os.UserHomeDir()
	if err != nil { return "", err }
	return filepath.Join(home, ".config", "orgm"), nil
}

// SaveBytes writes data to path creating directories
func SaveBytes(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil { return err }
	return os.WriteFile(path, data, 0644)
}

// Upload object to R2 bucket via S3 API
func r2S3Put(ctx context.Context, bucketKey, key string, body []byte) error {
	client, err := r2S3Client(ctx, bucketKey)
	if err != nil { return err }

	// Get the actual bucket name from viper
	bucketName := viper.GetString(fmt.Sprintf("cloudflare.r2_%s.CLOUDFLARE_R2_BUCKET", bucketKey))
	if bucketName == "" {
		// Fallback to old format
		bucketName = viper.GetString("cloudflare.bucket.orgm-privado.bucket")
		if bucketName == "" {
			bucketName = "orgm-privado" // default fallback
		}
	}

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &key,
		Body:   strings.NewReader(string(body)),
	})
	return err
}

// Download object from R2 bucket via S3 API
func r2S3Get(ctx context.Context, bucketKey, key string) ([]byte, error) {
	client, err := r2S3Client(ctx, bucketKey)
	if err != nil { return nil, err }

	// Get the actual bucket name from viper
	bucketName := viper.GetString(fmt.Sprintf("cloudflare.r2_%s.CLOUDFLARE_R2_BUCKET", bucketKey))
	if bucketName == "" {
		// Fallback to old format
		bucketName = viper.GetString("cloudflare.bucket.orgm-privado.bucket")
		if bucketName == "" {
			bucketName = "orgm-privado" // default fallback
		}
	}

	out, err := client.GetObject(ctx, &s3.GetObjectInput{Bucket: &bucketName, Key: &key})
	if err != nil { return nil, err }
	defer out.Body.Close()
	return io.ReadAll(out.Body)
}


