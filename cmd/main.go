package main

import (
	"context"
	"fiapx-worker/infra/database"
	"fiapx-worker/infra/service"
	"fiapx-worker/internal/usecase"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	log.Println("👷 Iniciando Worker...")

	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName:= os.Getenv("DB_NAME")
	awsEndpoint := os.Getenv("AWS_ENDPOINT")
	awsAccessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsRegion := os.Getenv("AWS_REGION")
	awsQueueURL := os.Getenv("AWS_QUEUE_URL")

	if dbHost == "" {
        dbHost = "localhost"
    }
	if dbUser == "" {
        dbUser = "user"
    }
	if dbPassword == "" {
        dbPassword = "password"
    }
	if dbName == "" {
        dbName = "fiapx_db"
    }
	if awsEndpoint == "" {
        awsEndpoint = "http://localhost:4566"
    }
	if awsAccessKeyID == "" {
        awsAccessKeyID = "teste"
    }
	if awsSecretAccessKey == "" {
        awsSecretAccessKey = "teste"
    }
	if awsRegion == "" {
        awsRegion = "us-east-1"
    }
	if awsQueueURL == "" {
        awsQueueURL = "http://localhost:4566/000000000000/video-processing-queue"
    }

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable", dbHost, dbUser, dbPassword, dbName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Error)})
	if err != nil {
		log.Fatalf("Erro Banco: %v", err)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: awsEndpoint}, nil
			})),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{AccessKeyID: awsAccessKeyID, SecretAccessKey: awsSecretAccessKey}, nil
		})),
	)
	if err != nil {
		log.Fatalf("Erro AWS Config: %v", err)
	}

	sqsClient := sqs.NewFromConfig(cfg)
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) { o.UsePathStyle = true })

	videoRepo := database.NewVideoRepository(db)
	storageService := service.NewStorageService(s3Client)
	mediaService := service.NewMediaService() // FFMPEG

	videoUC := usecase.NewVideoUseCase(videoRepo, storageService, mediaService)

	log.Println("🚀 Worker rodando e aguardando mensagens...")

	for {
		output, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(awsQueueURL),
			MaxNumberOfMessages: 1,
			WaitTimeSeconds:     10,
		})

		if err != nil {
			log.Printf("Erro SQS: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if len(output.Messages) > 0 {
			msg := output.Messages[0]
			videoID := *msg.Body

			err := videoUC.Execute(videoID)

			if err == nil || true {
				sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
					QueueUrl:      aws.String(awsQueueURL),
					ReceiptHandle: msg.ReceiptHandle,
				})
			}
		}
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}