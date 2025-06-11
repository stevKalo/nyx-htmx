package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func downloadFile(sess *session.Session) error {
	
	downloader := s3manager.NewDownloader(sess)
	
	f, err := os.Create("messages.txt")
	if err != nil {
		return fmt.Errorf("failed to create file 'messages.txt', %v", err)
	}

	defer f.Close()
	
	n, err := downloader.Download(f, &s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key:    aws.String("messages.txt"),
	})
	if err != nil {
		return fmt.Errorf("failed to download file, %v", err)
	}

	fmt.Printf("file downloaded, %d bytes\n", n)
	return nil
}

func uploadFile(sess *session.Session) error {
	
	uploader := s3manager.NewUploader(sess)

	f, err  := os.Open("messages.txt")
	if err != nil {
		return fmt.Errorf("failed to open file 'messages.txt', %v", err)
	}

	defer f.Close()

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key:    aws.String("messages.txt"),
		Body:   f,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file, %v", err)
	}

	fmt.Println("file uploaded to S3")
	return nil
}