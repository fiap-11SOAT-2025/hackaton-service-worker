package service

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type AWSClientFactory struct {
	Config aws.Config
}

// ADICIONAMOS O sessionToken AQUI NA ASSINATURA
func NewAWSClientFactory(ctx context.Context, region, endpoint, key, secret, sessionToken string) *AWSClientFactory {
	
	// 1. Configurações Base (Sempre carrega a Região)
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}

	// 2. Só sobrescreve o endpoint se não for vazio (ex: usando LocalStack)
	if endpoint != "" {
		opts = append(opts, config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: endpoint, SigningRegion: region}, nil
			})))
	}

	// 3. ⚠️ O SEGREDO AQUI: Só usa credenciais manuais se não for o valor "teste" ou vazio.
	// Se estiver a rodar no ECS da AWS, ele ignora este bloco e usa a LabRole automaticamente!
	if key != "" && key != "teste" {
		opts = append(opts, config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     key,
				SecretAccessKey: secret,
				SessionToken:    sessionToken,
			}, nil
		})))
	}

	// Carrega a configuração final com as opções escolhidas
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		log.Fatalf("Erro ao carregar configuração AWS: %v", err)
	}

	return &AWSClientFactory{Config: cfg}
}

func (f *AWSClientFactory) NewS3Client() *s3.Client {
	return s3.NewFromConfig(f.Config, func(o *s3.Options) {
		o.UsePathStyle = true
	})
}

func (f *AWSClientFactory) NewSQSClient() *sqs.Client {
	return sqs.NewFromConfig(f.Config)
}

func (f *AWSClientFactory) NewSESClient() *ses.Client {
	return ses.NewFromConfig(f.Config)
}

func (f *AWSClientFactory) NewSecretsManagerClient() *secretsmanager.Client {
	return secretsmanager.NewFromConfig(f.Config)
}

func (f *AWSClientFactory) NewSNSClient() *sns.Client {
	return sns.NewFromConfig(f.Config)
}