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

// NotifyError implements [usecase.Notifier].
func (s *NotificationService) NotifyError(videoID string, email string, errorMsg string) error {
	panic("unimplemented")
}

func NewNotificationService(client *sns.Client, topicArn string) *NotificationService {
	return &NotificationService{
		client:   client,
		topicArn: topicArn,
	}
}

func (s *NotificationService) SendNotification(email string, videoID string, status string) error {
	// O SNS envia para quem está subscrito no tópico, mas colocamos o e-mail do utilizador no texto para sabermos de quem é
	message := fmt.Sprintf("Olá,\n\nO processamento do vídeo (ID: %s) do utilizador %s foi concluído com o status: %s.\n\nSistema de Vídeos FIAP", videoID, email, status)
	subject := fmt.Sprintf("Atualização de Vídeo: %s", status)

	_, err := s.client.Publish(context.TODO(), &sns.PublishInput{
		Message:  aws.String(message),
		Subject:  aws.String(subject),
		TopicArn: aws.String(s.topicArn),
	})

	if err != nil {
		log.Printf("⚠️ Erro ao enviar notificação via SNS: %v", err)
		return err // Retorna o erro se falhar
	}

	log.Printf("✅ Notificação SNS enviada com sucesso para o Tópico!")
	return nil
}
