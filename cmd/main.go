package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"hackaton-service-worker/infra/database"
	"hackaton-service-worker/infra/service"
	"hackaton-service-worker/internal/entity" // Import necessário para o AutoMigrate
	"hackaton-service-worker/internal/usecase"
)

func main() {
	log.Println("👷 Iniciando Worker...")

	// 1. Variáveis AWS e SQS
	awsRegion := getEnv("AWS_REGION", "us-east-1")
	awsEndpoint := getEnv("AWS_ENDPOINT", "http://localhost:4566")
	awsAccessKeyID := getEnv("AWS_ACCESS_KEY_ID", "teste")
	awsSecretAccessKey := getEnv("AWS_SECRET_ACCESS_KEY", "teste")
	awsSessionToken := getEnv("AWS_SESSION_TOKEN", "") // (Obrigatória no Academy)
	awsQueueURL := getEnv("AWS_QUEUE_URL", "https://sqs.us-east-1.amazonaws.com/629000537837/video-processing-queue")
	
	// 👇 NOVA VARIÁVEL DO SNS AQUI
	awsSNSTopicARN := getEnv("AWS_SNS_TOPIC_ARN", "arn:aws:sns:us-east-1:000000000000:video-processing-notifications")

	// 2. Inicializa AWS PRIMEIRO
	awsFactory := service.NewAWSClientFactory(
		context.TODO(),
		awsRegion,
		awsEndpoint,
		awsAccessKeyID,
		awsSecretAccessKey,
		awsSessionToken,
	)

	// 3. Tenta buscar credenciais do Secrets Manager no formato da API
	secretName := getEnv("DB_SECRET_NAME", "database-credentials20260218011702627300000001")
	var dbHost, dbUser, dbPassword, dbName string

	secretsService := service.NewSecretsService(awsFactory.NewSecretsManagerClient())
	creds, err := secretsService.GetDatabaseCredentials(context.TODO(), secretName)
	
	if err == nil {
		fmt.Println("✅ Credenciais carregadas do AWS Secrets Manager")
		dbHost = creds.Host
		dbUser = creds.Username
		dbPassword = creds.Password
		dbName = creds.DbName
	} else {
		fmt.Printf("⚠️ Erro ao acessar secret (%v). Usando variáveis locais.\n", err)
		// Fallback para variáveis de ambiente (útil se o Secrets Manager falhar ou rodando local)
		dbHost = getEnv("DB_HOST", "localhost")
		dbUser = getEnv("DB_USER", "user")
		dbPassword = getEnv("DB_PASSWORD", "password")
		dbName = getEnv("DB_NAME", "fiapx_db")
	}

	// 4. Inicialização do Banco de Dados
	db := database.SetupDatabase(dbHost, dbUser, dbPassword, dbName)
	if db == nil {
		// Se o banco não conectar, não adianta continuar. Encerra com erro.
		panic("❌ Falha crítica: Banco de dados não inicializado.")
	}
	
	// O Worker só precisa migrar a entidade Video (User fica com a API)
	db.AutoMigrate(&entity.Video{})

	// 5. Inicialização dos Serviços e Repositórios
	videoRepo := database.NewVideoRepository(db)
	storageService := service.NewStorageService(awsFactory.NewS3Client())
	mediaService := service.NewMediaService()
	
	// 👇 INICIALIZANDO O SERVIÇO DE NOTIFICAÇÃO COM O CLIENTE SNS
	notificationService := service.NewNotificationService(awsFactory.NewSNSClient(), awsSNSTopicARN)

	videoUC := usecase.NewVideoUseCase(videoRepo, storageService, mediaService, notificationService)

	// 6. Inicia o Worker escutando a fila SQS
	worker := service.NewQueueService(awsFactory.NewSQSClient(), awsQueueURL, videoUC)
	worker.Start(context.Background())
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}