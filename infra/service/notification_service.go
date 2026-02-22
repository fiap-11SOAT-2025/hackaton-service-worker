package service

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type NotificationService struct {
	client   *sns.Client
	topicArn string
}

func NewNotificationService(client *sns.Client, topicArn string) *NotificationService {
	return &NotificationService{
		client:   client,
		topicArn: topicArn,
	}
}

func (s *NotificationService) SendNotification(email string, videoID string, status string) error {
	message := fmt.Sprintf("Olá,\n\nO processamento do vídeo (ID: %s) do utilizador %s foi concluído com o status: %s.\n\nSistema de Vídeos FIAP", videoID, email, status)
	subject := fmt.Sprintf("Atualização de Vídeo: %s", status)

	return s.publishToSNS(subject, message)
}

func (s *NotificationService) NotifyError(videoID string, email string, errorMsg string) error {
	message := fmt.Sprintf("Olá,\n\nInfelizmente ocorreu um erro ao processar o seu vídeo (ID: %s).\nErro: %s\n\nSistema de Vídeos FIAP", videoID, errorMsg)
	subject := "Falha no Processamento do Vídeo"

	return s.publishToSNS(subject, message)
}

func (s *NotificationService) publishToSNS(subject, message string) error {
	_, err := s.client.Publish(context.TODO(), &sns.PublishInput{
		Message:  aws.String(message),
		Subject:  aws.String(subject),
		TopicArn: aws.String(s.topicArn),
	})

	if err != nil {
		log.Printf("⚠️ Erro ao enviar notificação via SNS: %v", err)
		return err
	}

	log.Printf("✅ Notificação SNS enviada com sucesso para o Tópico!")
	return nil
}