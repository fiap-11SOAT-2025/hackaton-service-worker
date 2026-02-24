package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"hackaton-service-worker/infra/database"
	"hackaton-service-worker/infra/service"
	"hackaton-service-worker/internal/entity"
	"hackaton-service-worker/internal/usecase"
)

func main() {
	log.Println("👷 Iniciando Worker...")

	awsRegion := getEnv("AWS_REGION", "us-east-1")
	awsEndpoint := getEnv("AWS_ENDPOINT", "http://localhost:4566")
	awsAccessKeyID := getEnv("AWS_ACCESS_KEY_ID", "teste")
	awsSecretAccessKey := getEnv("AWS_SECRET_ACCESS_KEY", "teste")
	awsSessionToken := getEnv("AWS_SESSION_TOKEN", "")
	awsQueueURL := getEnv("AWS_QUEUE_URL", "https://sqs.us-east-1.amazonaws.com/225022246839/video-processing-queue")
	
	awsSNSTopicARN := getEnv("AWS_SNS_TOPIC_ARN", "arn:aws:sns:us-east-1:000000000000:video-processing-notifications")

	awsFactory := service.NewAWSClientFactory(
		context.TODO(),
		awsRegion,
		awsEndpoint,
		awsAccessKeyID,
		awsSecretAccessKey,
		awsSessionToken,
	)

	secretName := getEnv("DB_SECRET_NAME", "database-credentials20260224231056274100000001")
	var dbHost, dbUser, dbPassword, dbName, dbSslmode string

	secretsService := service.NewSecretsService(awsFactory.NewSecretsManagerClient())
	creds, err := secretsService.GetDatabaseCredentials(context.TODO(), secretName)
	
	if err == nil {
		fmt.Println("✅ Credenciais carregadas do AWS Secrets Manager")
		dbHost = creds.Host
		dbUser = creds.Username
		dbPassword = creds.Password
		dbName = creds.Name
		dbSslmode = creds.Sslmode
	} else {
		fmt.Printf("⚠️ Erro ao acessar secret (%v). Usando variáveis locais.\n", err)
		dbHost = getEnv("DB_HOST", "localhost")
		dbUser = getEnv("DB_USER", "user")
		dbPassword = getEnv("DB_PASSWORD", "password")
		dbName = getEnv("DB_NAME", "fiapx_db")
		dbSslmode = getEnv("DB_SSL_MODE", "disable")
	}

	// Certifique-se que seu postgres.go está com sslmode=require para AWS e sslmode=disable para Local 
	db := database.SetupDatabase(dbHost, dbUser, dbPassword, dbName, dbSslmode)
	if db == nil {
		panic("❌ Falha crítica: Banco de dados não inicializado.")
	}
	
	db.AutoMigrate(&entity.Video{})

	videoRepo := database.NewVideoRepository(db)
	storageService := service.NewStorageService(awsFactory.NewS3Client())
	mediaService := service.NewMediaService()
	
	notificationService := service.NewNotificationService(awsFactory.NewSNSClient(), awsSNSTopicARN)

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