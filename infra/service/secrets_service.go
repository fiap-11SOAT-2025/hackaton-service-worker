package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type DatabaseCredentials struct {
	Username string `json:"DB_USERNAME"`
	Password string `json:"DB_PASSWORD"`
	Host     string `json:"DB_HOST"`
	DbName   string `json:"DB_NAME"`
	Port     string `json:"DB_PORT"`
}

type SecretsService struct {
	client *secretsmanager.Client
}

func NewSecretsService(client *secretsmanager.Client) *SecretsService {
	return &SecretsService{client: client}
}

func (s *SecretsService) GetDatabaseCredentials(ctx context.Context, secretName string) (*DatabaseCredentials, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := s.client.GetSecretValue(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar segredo %s: %w", secretName, err)
	}

	var creds DatabaseCredentials
	if err := json.Unmarshal([]byte(*result.SecretString), &creds); err != nil {
		return nil, fmt.Errorf("erro ao converter json do segredo: %w", err)
	}

	return &creds, nil
}