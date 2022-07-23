package s3

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/FideTechSolutions/uberContainer/config"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Client struct {
	config.S3Config
}

func NewClient(conf config.S3Config) *S3Client {
	return &S3Client{
		conf,
	}
}

// Download will download the file to temp file store
func (s *S3Client) Download(file string) (io.Reader, error) {
	minioClient, err := minio.New(s.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.AccessKey, s.AccessSecret, ""),
		Secure: s.UseSSL,
		Region: s.Region,
	})
	if err != nil {
		return nil, err
	}

	file = strings.TrimPrefix(file, "/")

	object, err := minioClient.GetObject(
		context.Background(),
		s.Bucket,
		file,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, err
	}

	return object, nil
}

func (s *S3Client) GetDownloadLink(file string) (string, error) {
	expire := 5 * time.Minute

	minioClient, err := minio.New(s.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.AccessKey, s.AccessSecret, ""),
		Secure: s.UseSSL,
		Region: s.Region,
	})
	if err != nil {
		return "", err
	}

	file = strings.TrimPrefix(file, "/")

	url, err := minioClient.PresignedGetObject(context.Background(), s.Bucket, file, expire, make(url.Values))
	if err != nil {
		return "", err
	}

	return url.String(), nil
}

// Upload will upload the file from given file path
func (s *S3Client) Upload(objectPath string, filePath string, contentType string) error {
	minioClient, err := minio.New(s.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.AccessKey, s.AccessSecret, ""),
		Secure: s.UseSSL,
		Region: s.Region,
	})
	if err != nil {
		return err
	}

	objectPath = strings.TrimPrefix(objectPath, "/")

	_, err = minioClient.FPutObject(
		context.Background(),
		s.Bucket,
		objectPath,
		filePath,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// UploadFromReader will upload the file from given file path
func (s *S3Client) UploadFromReader(objectPath string, contentType string, file io.Reader, fileSize int64) error {
	minioClient, err := minio.New(s.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.AccessKey, s.AccessSecret, ""),
		Secure: s.UseSSL,
		Region: s.Region,
	})
	if err != nil {
		return err
	}

	objectPath = strings.TrimPrefix(objectPath, "/")

	_, err = minioClient.PutObject(context.Background(), s.Bucket, objectPath, file, fileSize, minio.PutObjectOptions{
		ContentType: contentType,
	})

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// Delete permanentely permanentely destroys an object specified
func (s *S3Client) Delete(file string) error {
	minioClient, err := minio.New(s.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.AccessKey, s.AccessSecret, ""),
		Secure: s.UseSSL,
		Region: s.Region,
	})
	if err != nil {
		return err
	}

	file = strings.TrimPrefix(file, "/")

	// removes the bucket name from the Path if it exists
	objectPrefix := strings.TrimPrefix(file, s.Bucket)

	// chan of objects withing the deployment object
	objectsCh := make(chan string)

	//send object names thata are to be removed to objectsCh
	go func() {
		defer close(objectsCh)
		recursive := true

		for object := range minioClient.ListObjects(
			context.Background(),
			s.Bucket,
			minio.ListObjectsOptions{
				Prefix:    objectPrefix,
				Recursive: recursive,
			},
		) {
			if object.Err != nil {
				log.Println(object.Err)
			}

			objectsCh <- object.Key
		}
	}()

	log.Println("permanentely deleting all the objects inside folder (if folder)")
	for objName := range objectsCh {
		err := minioClient.RemoveObject(
			context.Background(),
			s.Bucket,
			objName,
			minio.RemoveObjectOptions{})
		if err != nil {
			return err
		}

	}

	log.Println("permanentely deleting the object itself")
	return minioClient.RemoveObject(
		context.Background(),
		s.Bucket,
		file,
		minio.RemoveObjectOptions{})

}
