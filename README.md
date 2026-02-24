# FIAP X - Video Processor Worker

O Worker é um microsserviço especializado em processamento pesado de média. Ele consome pedidos de uma fila SQS, extrai frames de vídeos utilizando FFMPEG e envia notificações de estado via Amazon SNS.

## ⚙️ Fluxo de Trabalho
1. **Consumo (SQS)**: Monitoriza continuamente a fila `video-processing-queue`.
2. **Estado (Base de Dados)**: Altera o status do vídeo para `PROCESSING` no PostgreSQL.
3. **Download (S3)**: Recupera o ficheiro original do Bucket de entrada.
4. **Extração (FFMPEG)**: Processa o vídeo para capturar 1 frame por segundo.
5. **Compactação**: Reúne as imagens geradas num ficheiro ZIP.
6. **Upload (S3)**: Envia o resultado final para o diretório `outputs/`.
7. **Estado (Base de Dados)**: Altera o status do vídeo ao fim do processo no PostgreSQL.
8. **Notificação (SNS)**: Dispara um alerta de sucesso ou erro para o tópico configurado.

## 🚀 Tecnologias e Recursos
- **Linguagem**: Go 1.24
- **Processamento de Vídeo**: FFMPEG (instalado via Alpine)
- **Infraestrutura Cloud (AWS SDK v2)**:
  - **SQS**: Gestão de fila de mensagens.
  - **S3**: Armazenamento de objetos.
  - **SNS**: Sistema de notificações (Pub/Sub).
  - **Secrets Manager**: Recuperação segura de credenciais de base de dados.
- **ORM**: GORM (PostgreSQL) com suporte a SSL.

## 📦 Variáveis de Ambiente Principais
O worker está configurado para suportar ambientes locais (LocalStack) e AWS real (incluindo AWS Academy via `AWS_SESSION_TOKEN`).

| Variável | Descrição | Exemplo |
| :--- | :--- | :--- |
| `AWS_REGION` | Região da AWS | `us-east-1` |
| `AWS_ENDPOINT` | URL do mock (se usar LocalStack) | `http://localstack:4566` |
| `AWS_SNS_TOPIC_ARN` | ARN do tópico para notificações | `arn:aws:sns:us-east-1:000...` |
| `DB_SECRET_NAME` | Nome do segredo no Secrets Manager | `db-credentials` |
| `DB_SSL_MODE` | Modo de segurança da ligação DB | `require` ou `disable` |

## 🛠️ Execução via Docker
Certifique-se de que o PostgreSQL e o LocalStack (ou AWS) estão acessíveis. Para mais detalhes acesse [hackaton-service-api](https://github.com/fiap-11SOAT-2025/hackaton-service-api).

```bash
# Construir a imagem
docker build -t hackaton-service-worker .
```

```bash
# Executar localmente (exemplo)
docker run --env-file .env hackaton-service-worker
```

## 🧪 Testes

Para garantir a integridade da lógica de processamento de vídeos:

```bash
go test -coverprofile=coverage.out ./internal/usecase/...
```
Para gerar em html:

```bash
go tool cover -html=coverage.out
```

