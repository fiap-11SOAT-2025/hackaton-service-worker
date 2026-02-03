package service

import (
	"context"
	"encoding/json"
	"hackaton-service-worker/internal/usecase"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSMessage struct {
	VideoID string `json:"video_id"`
	Email   string `json:"email"`
}

type QueueService struct {
	SQSClient *sqs.Client
	QueueURL  string
	VideoUC   *usecase.VideoUseCase
}

func NewQueueService(client *sqs.Client, queueURL string, videoUC *usecase.VideoUseCase) *QueueService {
	return &QueueService{
		SQSClient: client,
		QueueURL:  queueURL,
		VideoUC:   videoUC,
	}
}

func (s *QueueService) Start(ctx context.Context) {
	log.Println("🚀 Worker consumindo da fila...")

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 Parando consumo da fila...")
			return
		default:
			s.consume(ctx)
		}
	}
}

func (s *QueueService) consume(ctx context.Context) {
	output, err := s.SQSClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.QueueURL),
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     10,
	})

	if err != nil {
		log.Printf("Erro SQS: %v", err)
		time.Sleep(5 * time.Second)
		return
	}

	for _, msg := range output.Messages {
		var body SQSMessage
		if err := json.Unmarshal([]byte(*msg.Body), &body); err != nil {
			log.Printf("Erro JSON inválido: %v", err)
			s.deleteMessage(ctx, msg.ReceiptHandle)
			continue
		}

		err := s.VideoUC.Execute(body.VideoID, body.Email)

		if err == nil {
			s.deleteMessage(ctx, msg.ReceiptHandle)
		}
	}
}

func (s *QueueService) deleteMessage(ctx context.Context, receiptHandle *string) {
	_, err := s.SQSClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.QueueURL),
		ReceiptHandle: receiptHandle,
	})
	if err != nil {
		log.Printf("Erro ao deletar mensagem: %v", err)
	}
}