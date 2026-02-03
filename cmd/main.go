package main

import (
	"context"
	"hackaton-service-worker/infra/database"
	"hackaton-service-worker/infra/service"
	"hackaton-service-worker/internal/usecase"
	"log"
	"os"
)

func main() {
	log.Println("👷 Iniciando Worker...")

	dbHost := getEnv("DB_HOST", "localhost")
	dbUser := getEnv("DB_USER", "user")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "fiapx_db")
	
	awsRegion := getEnv("AWS_REGION", "us-east-1")
	awsEndpoint := getEnv("AWS_ENDPOINT", "http://localhost:4566")
	awsAccessKeyID := getEnv("AWS_ACCESS_KEY_ID", "teste")
	awsSecretAccessKey := getEnv("AWS_SECRET_ACCESS_KEY", "teste")
	awsQueueURL := getEnv("AWS_QUEUE_URL", "http://localhost:4566/000000000000/video-processing-queue")
	emailIdentity := getEnv("EMAIL_IDENTITY", "no-reply@fiapx.com")	

	db := database.SetupDatabase(dbHost, dbUser, dbPassword, dbName)

	awsFactory := service.NewAWSClientFactory(
        context.TODO(),
        awsRegion,
        awsEndpoint,
        awsAccessKeyID,
        awsSecretAccessKey,
    )

	videoRepo := database.NewVideoRepository(db)
	storageService := service.NewStorageService(awsFactory.NewS3Client())
	notificationService := service.NewNotificationService(awsFactory.NewSESClient(), emailIdentity)
	mediaService := service.NewMediaService()

	videoUC := usecase.NewVideoUseCase(videoRepo, storageService, mediaService, notificationService)

	worker := service.NewQueueService(awsFactory.NewSQSClient(), awsQueueURL, videoUC)

    worker.Start(context.Background())
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}