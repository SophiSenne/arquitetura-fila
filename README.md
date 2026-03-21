# Sensor Telemetry API

Sistema de ingestão de telemetria para dispositivos embarcados industriais. O backend recebe leituras de sensores via HTTP e as encaminha para uma fila RabbitMQ, onde um consumidor as processa e persiste em um banco de dados PostgreSQL.

## Visão Geral

A aplicação resolve um desafio de escalabilidade no monitoramento industrial: dispositivos embarcados enviam leituras periódicas de sensores (temperatura, umidade, presença, vibração, luminosidade e nível de reservatórios) para um backend central. Para suportar alto volume de requisições simultâneas sem gargalos, o processamento é assíncrono, ou seja, o endpoint apenas enfileira a mensagem e retorna imediatamente, enquanto um consumidor independente persiste os dados no banco.


## Arquitetura

```
┌─────────────────┐     HTTP POST      ┌─────────────────┐
│ Dispositivo      │ ─────────────────► │   Backend Go     │
│ Embarcado        │                    │   (porta 8080)   │
└─────────────────┘                    └────────┬────────┘
                                                │ AMQP Publish
                                                ▼
                                       ┌─────────────────┐
                                       │    RabbitMQ      │
                                       │  fila: SENSOR.DATA│
                                       │  (porta 5672)    │
                                       └────────┬────────┘
                                                │ AMQP Consume
                                                ▼
                                       ┌─────────────────┐
                                       │   Middleware Go  │
                                       │   (Consumidor)   │
                                       └────────┬────────┘
                                                │ SQL Insert
                                                ▼
                                       ┌─────────────────┐
                                       │   PostgreSQL     │
                                       │   (porta 5433)   │
                                       └─────────────────┘
```

O fluxo completo funciona da seguinte forma:

1. O dispositivo embarcado envia um `POST /sensorData` com os dados da leitura
2. O Backend valida o JSON e publica a mensagem na fila `SENSOR.DATA` do RabbitMQ
3. O Middleware consome as mensagens da fila e as persiste no PostgreSQL
4. Em caso de falha na persistência, a mensagem é reenfileirada (`Nack` com `requeue=true`)
5. O middleware se reconecta automaticamente ao RabbitMQ em caso de queda (backoff de 5s)

## Estrutura do Projeto

```
.
├── back-end/                  # API HTTP (produtor RabbitMQ)
│   ├── handler/
│   │   └── sensor.go          # Handler dos endpoints HTTP
│   ├── messaging/
│   │   └── rabbitmq.go        # Publicação de mensagens no RabbitMQ
│   ├── model/
│   │   └── sensor.go          # Struct SensorMessage
│   ├── router/
│   │   └── router.go          # Configuração das rotas Gin
│   ├── main.go
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
│
├── middleware/                # Consumidor RabbitMQ + persistência
│   ├── cmd/middleware/
│   │   └── main.go            # Ponto de entrada, conexões DB e RabbitMQ
│   ├── consumer/
│   │   └── consumer_job.go    # Loop de consumo
│   ├── data/
│   │   └── sensor.go          # Repositório PostgreSQL
│   ├── model/
│   │   └── sensor.go          # Struct SensorMessage
│   ├── unit-test/
│   │   └── sensor_test.go     # Testes unitários
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
│
├── database/                  # PostgreSQL
│   ├── init/
│   │   └── init.sql           # Schema + seed de dados
│   └── Dockerfile
│
├── rabbitMQ/                  # RabbitMQ
│   ├── definitions.json       # Usuários, vhosts, permissões e fila pré-criada
│   ├── rabbitmq.conf          # Carrega definitions.json na inicialização
│   └── Dockerfile
│
├── load-test/                 # Teste de carga k6
│   ├── k6-test.ts             # Script TypeScript do k6
│   └── Dockerfile             # Build TypeScript + execução k6
│
└── docker-compose.yaml        # Orquestração de todos os serviços
```
## Modelo de Dados

O banco é estruturado para separar leituras analógicas (valores contínuos como temperatura) de discretas (valores inteiros como presença/ausência), permitindo otimizações de armazenamento e consulta distintas para cada tipo.

### Diagrama de Entidades

```
sensor_types          devices
─────────────         ─────────────────────
id   (PK)             id          (PK UUID)
name (UNIQUE)         serial      (UNIQUE)
unit                  description
                      active
                      created_at

analog_readings                 discrete_readings
──────────────────────────      ──────────────────────────
id           (PK BIGSERIAL)     id           (PK BIGSERIAL)
device_id    (FK → devices)     device_id    (FK → devices)
sensor_type_id (FK → sensor_types) sensor_type_id (FK → sensor_types)
value        NUMERIC(10,4)      value        INTEGER
collected_at TIMESTAMPTZ        collected_at TIMESTAMPTZ
saved_at     TIMESTAMPTZ        saved_at     TIMESTAMPTZ
```

### Tipos de Sensores (seed)

| Nome | Unidade | Tipo de Leitura |
|---|---|---|
| temperatura | °C | analógica |
| umidade | % | analógica |
| presença | — | discreta (0/1) |
| vibração | mm/s | analógica |
| luminosidade | lux | analógica |
| nível_reservatório | % | analógica |

### Dispositivos Pré-cadastrados (seed)

| Serial | Descrição | Ativo |
|---|---|---|
| SN-TH-001 | Temperatura/Umidade — Sala de servidores A | ✅ |
| SN-TH-002 | Temperatura/Umidade — Sala de servidores B | ✅ |
| SN-PIR-001 | Presença PIR — Corredor principal | ✅ |
| SN-VIB-001 | Vibração — Compressor #1 | ✅ |
| SN-LUX-001 | Luminosidade — Área de produção | ✅ |
| SN-NIV-001 | Nível — Reservatório de água tratada | ✅ |
| SN-TH-003 | Temperatura — Câmara fria | ❌ (inativo) |

## API

### `GET /`

Verifica se a API está no ar.

**Resposta `200 OK`:**
```json
{ "success": "API running" }
```

### `POST /sensorData`

Recebe uma leitura de sensor e a enfileira no RabbitMQ para processamento assíncrono.

**Headers:**
```
Content-Type: application/json
```

**Body:**
```json
{
  "idSensor":   "SN-TH-001",
  "timestamp":  "2024-06-01T12:00:00Z",
  "sensorType": "temperatura",
  "readType":   "analog",
  "value":      23.5
}
```

| Campo | Tipo | Descrição |
|---|---|---|
| `idSensor` | string | Serial do dispositivo (deve existir na tabela `devices`) |
| `timestamp` | string | Data/hora da coleta em formato RFC 3339 |
| `sensorType` | string | Tipo do sensor (deve existir na tabela `sensor_types`) |
| `readType` | string | `"analog"` ou `"discrete"` |
| `value` | number | Valor da leitura |

**Respostas:**

| Status | Descrição |
|---|---|
| `200 OK` | Mensagem enfileirada com sucesso |
| `400 Bad Request` | JSON inválido ou campos ausentes |
| `500 Internal Server Error` | Falha na conexão com o RabbitMQ |

**Resposta `200 OK`:**
```json
{ "success": "Message sent to queue" }
```

**Resposta `400 Bad Request`:**
```json
{ "error": "Invalid JSON" }
```

---

## Como Executar

### Pré-requisitos

- [Docker](https://docs.docker.com/get-docker/) instalado
- [Docker Compose](https://docs.docker.com/compose/install/) instalado

### 1. Subir todos os serviços

```bash
docker compose up --build
```

Os serviços estarão disponíveis em:

| Serviço | Endereço |
|---|---|
| API Backend | http://localhost:8080 |
| RabbitMQ Management UI | http://localhost:15672 (usuário: `admin`, senha: `admin`) |
| PostgreSQL | localhost:5433 (usuário: `admin`, senha: `admin`, banco: `database`) |

### 2. Testar o endpoint

```bash
curl -X POST http://localhost:8080/sensorData \
  -H "Content-Type: application/json" \
  -d '{
    "idSensor":   "SN-TH-001",
    "timestamp":  "2024-06-01T12:00:00Z",
    "sensorType": "temperatura",
    "readType":   "analog",
    "value":      23.5
  }'
```

### 3. Derrubar os serviços

```bash
docker compose down
```

Para remover volumes (apaga dados do banco e do RabbitMQ):

```bash
docker compose down -v
```

## Testes de Carga

O teste de carga é executado com **k6** dentro de um container Docker, simulando múltiplos dispositivos enviando leituras simultâneas ao backend.

### Cenário

O teste utiliza uma rampa progressiva de usuários virtuais (VUs):

```
VUs
50 │                    ████████████
   │                 ███            ███
10 │    █████████████                  
   │ ███                                ███
 0 └────────────────────────────────────────► tempo
     0s  30s    90s  120s         180s 210s
```

### Thresholds (critérios de sucesso)

| Métrica | Critério |
|---|---|
| `http_req_duration` (p95) | < 500ms |
| `sensor_error_rate` | < 1% |
| `sensor_request_duration` (mediana) | < 200ms |

### Métricas customizadas

- `sensor_error_rate` — taxa de requisições com falha
- `sensor_request_duration` — duração das requisições (Trend)
- `sensor_total_requests` — contador total de requisições enviadas

### Executar o teste de carga

Com todos os serviços em execução, execute em um terminal separado:

```bash
docker compose --profile loadtest up k6
```

O k6 simula os 7 perfis de sensores cadastrados no seed, gerando payloads realistas com valores dentro das faixas esperadas para cada tipo de sensor.

## Testes Unitários

Os testes unitários estão no módulo `middleware`, cobrindo a lógica de negócio sem dependências externas (banco de dados ou RabbitMQ).

### O que é testado

- Serialização e desserialização JSON do `SensorMessage`
- Validação dos campos `readType` (`analog` / `discrete`)
- Parsing de timestamps RFC 3339 (válidos e inválidos)
- Comportamento do repositório fake (`fakeRepo`)
- Lógica de `handleMessage` (parse → save → ack/nack)
- Processamento em lote (`SaveReadingsBatch`) e preservação de ordem
- Casos de borda: valores zero, negativos e muito grandes
- Geração de DSNs para PostgreSQL e RabbitMQ

### Executar os testes

```bash
cd middleware
go test ./unit-test/... -v
```

Ou, usando Docker diretamente no container do middleware:

```bash
docker compose run --rm middleware go test ./unit-test/... -v
```

## Variáveis de Ambiente

### Backend

| Variável | Padrão | Descrição |
|---|---|---|
| `RABBITMQ_HOST` | — | Hostname do RabbitMQ |
| `RABBITMQ_PORT` | — | Porta AMQP (geralmente `5672`) |
| `RABBITMQ_USER` | — | Usuário do RabbitMQ |
| `RABBITMQ_PASSWORD` | — | Senha do RabbitMQ |

### Middleware

| Variável | Padrão | Descrição |
|---|---|---|
| `DB_HOST` | `localhost` | Hostname do PostgreSQL |
| `DB_PORT` | `5432` | Porta do PostgreSQL |
| `DB_USER` | `admin` | Usuário do banco |
| `DB_PASSWORD` | `admin` | Senha do banco |
| `DB_NAME` | `database` | Nome do banco |
| `RABBITMQ_HOST` | `localhost` | Hostname do RabbitMQ |
| `RABBITMQ_PORT` | `5672` | Porta AMQP |
| `RABBITMQ_USER` | `admin` | Usuário do RabbitMQ |
| `RABBITMQ_PASSWORD` | `admin` | Senha do RabbitMQ |