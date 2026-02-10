package uploader

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/mlogclub/simple/common/strs"

	"bbs-go/internal/models/dto"
)

// S3 obs
type S3ObsUploader struct {
	m          sync.Mutex
	client     *s3.Client
	currentCfg dto.UploadConfig
}

func (u *S3ObsUploader) PutImage(cfg *dto.UploadConfig, data []byte, contentType string) (string, error) {
	if strs.IsBlank(contentType) {
		contentType = "image/jpeg"
	}
	key := generateImageKey(data, contentType)
	return u.PutObject(cfg, key, data, contentType)
}

func (u *S3ObsUploader) PutObject(cfg *dto.UploadConfig, key string, data []byte, contentType string) (string, error) {
	if err := u.initClient(cfg); err != nil {
		return "", nil
	}
	opt := &s3.PutObjectInput{
		Bucket:      &cfg.S3Obs.Bucket,
		Key:         &key,
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
		ACL:         types.ObjectCannedACLPublicRead,
	}
	if _, err := u.client.PutObject(context.Background(), opt); err != nil {
		return "", err
	}
	baseURL := cfg.S3Obs.BaseURL
	if strs.IsBlank(baseURL) {
		protocol := "http"
		if cfg.S3Obs.UseSSL {
			protocol = "https"
		}
		baseURL = fmt.Sprintf("%s://%s.%s", protocol, cfg.S3Obs.Bucket, cfg.S3Obs.Endpoint)
	}
	return fmt.Sprintf("%s/%s", baseURL, key), nil
}

func (u *S3ObsUploader) CopyImage(cfg *dto.UploadConfig, originUrl string) (string, error) {
	data, contentType, err := download(originUrl)
	if err != nil {
		return "", err
	}
	return u.PutImage(cfg, data, contentType)
}

func (u *S3ObsUploader) initClient(cfg *dto.UploadConfig) error {
	if !u.isCfgChange(cfg) {
		return nil
	}

	u.m.Lock()
	defer u.m.Unlock()

	if cfg != nil {
		defaultConfig, err := config.LoadDefaultConfig(context.Background(),
			config.WithRegion(cfg.S3Obs.Region),
			config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(
					cfg.S3Obs.AccessKey,
					cfg.S3Obs.SecretKey,
					"",
				),
			),
		)
		if err != nil {
			slog.Error("init s3 obs client error", slog.Any("err", err))
			return err
		}
		endpoint := fmt.Sprintf("http://%s", cfg.S3Obs.Endpoint)
		if cfg.S3Obs.UseSSL {
			endpoint = fmt.Sprintf("https://%s", endpoint)
		}
		defaultConfig.BaseEndpoint = &endpoint
		defaultConfig.HTTPClient = &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: true,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		u.client = s3.NewFromConfig(defaultConfig, func(o *s3.Options) {
			o.BaseEndpoint = &endpoint
			o.UsePathStyle = true
		})
		u.currentCfg = *cfg
	}
	return nil
}

func (u *S3ObsUploader) isCfgChange(cfg *dto.UploadConfig) bool {
	if cfg == nil || u.client == nil {
		return true
	}
	if u.currentCfg.S3Obs.Endpoint != cfg.S3Obs.Endpoint ||
		u.currentCfg.S3Obs.Bucket != cfg.S3Obs.Bucket ||
		u.currentCfg.S3Obs.AccessKey != cfg.S3Obs.AccessKey ||
		u.currentCfg.S3Obs.SecretKey != cfg.S3Obs.SecretKey ||
		u.currentCfg.S3Obs.UseSSL != cfg.S3Obs.UseSSL {
		return true
	}
	return false
}
