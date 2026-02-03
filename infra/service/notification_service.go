package service

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type NotificationService struct {
	SESClient *ses.Client
	Sender    string
}

func NewNotificationService(client *ses.Client, sender string) *NotificationService {
	return &NotificationService{SESClient: client, Sender: sender}
}

func (n *NotificationService) NotifyError(videoID, email, errorMsg string) error {
	subject := "FIAP X - Falha no Processamento"
	body := fmt.Sprintf("Olá,\n\nInfelizmente o processamento do vídeo %s falhou.\n\nErro: %s\n\nAtenciosamente,\nEquipe FIAP X", videoID, errorMsg)

	input := &ses.SendEmailInput{
		Source: aws.String(n.Sender),
		Destination: &types.Destination{
			ToAddresses: []string{email},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data: aws.String(subject),
			},
			Body: &types.Body{
				Text: &types.Content{
					Data: aws.String(body),
				},
			},
		},
	}

	_, err := n.SESClient.SendEmail(context.TODO(), input)
	if err != nil {
		log.Printf("⚠️ Erro ao enviar e-mail via SES para %s: %v", email, err)
		return err
	}

	log.Printf("📧 E-mail enviado com sucesso via SES para: %s", email)
	return nil
}