# Flight Award Integration - Gemini Response Processing

Este documento explica como usar o sistema de integração entre as respostas do Gemini e as tabelas de Flight Award.

## Visão Geral

O sistema permite converter e salvar dados de respostas do Gemini API (que processa imagens de reservas de voos) para as tabelas estruturadas do banco de dados PostgreSQL.

## Estruturas de Dados

### GeminiResponse

Representa a resposta estruturada do Gemini API quando processa imagens de reservas de voos:

```go
type GeminiResponse struct {
    Origin          GeminiOrigin           `json:"origin"`
    Destination     GeminiDestination      `json:"destination"`
    Airline         string                 `json:"airline"`
    ServiceClass    string                 `json:"serviceClass"`
    LoyaltyPrograms []GeminiLoyaltyProgram `json:"loyaltyPrograms"`
    Availability    GeminiAvailability     `json:"availability"`
    Connections     []GeminiConnection     `json:"connections"`
}
```

### FlightAward (Tabelas do Banco)

Representa a estrutura normalizada no banco de dados com todas as tabelas relacionadas:

- `FlightAward` - Dados principais do voo
- `Airport` - Aeroportos (origem, destino, conexões)
- `Airline` - Companhias aéreas
- `LoyaltyProgram` - Programas de fidelidade
- `FlightAwardAvailableDate` - Datas disponíveis
- `FlightAwardProgramCost` - Custos por programa de fidelidade
- `FlightConnection` - Conexões do voo

## Serviços Disponíveis

### FlightAwardService

O serviço principal que gerencia a conversão e persistência dos dados:

```go
service := NewFlightAwardService(db)
```

#### Métodos Principais:

1. **SaveGeminiResponse(geminiResp \*GeminiResponse)**

   - Converte e salva uma resposta do Gemini no banco de dados
   - Retorna o FlightAward criado
   - Usa transações para garantir consistência

2. **GetFlightAwardByID(id uuid.UUID)**

   - Recupera um FlightAward por ID com todos os dados relacionados

3. **GetFlightAwardsByRoute(originCode, destCode string)**
   - Busca FlightAwards por rota (origem -> destino)

## Endpoints da API

### 1. Processar Resposta do Gemini

```
POST /api/flight-award/process-gemini
```

**Request Body:**

```json
{
  "geminiResponse": {
    "origin": {
      "city": "São Paulo",
      "airportCode": "GRU"
    },
    "destination": {
      "city": "Nova York",
      "airportCode": "JFK"
    },
    "airline": "LATAM",
    "serviceClass": "Business",
    "loyaltyPrograms": [
      {
        "name": "LATAM Pass",
        "miles": 50000,
        "fees": {
          "value": 150.0,
          "currency": "USD"
        }
      }
    ],
    "availability": {
      "searchDate": "2024-01-15",
      "availableDates": ["2024-02-01", "2024-02-05"]
    },
    "connections": [
      {
        "city": "Brasília",
        "airportCode": "BSB"
      }
    ]
  }
}
```

**Response:**

```json
{
  "success": true,
  "flightAward": {
    "id": "uuid-here",
    "originAirportCode": "GRU",
    "destinationAirportCode": "JFK"
    // ... outros campos
  },
  "message": "Flight award data saved successfully"
}
```

### 2. Buscar Flight Award por ID

```
GET /api/flight-award/{id}
```

### 3. Buscar Flight Awards por Rota

```
GET /api/flight-award/route?origin=GRU&destination=JFK
```

## Funcionalidades Implementadas

### Conversão Automática de Dados

1. **Aeroportos**: Cria automaticamente aeroportos se não existirem
2. **Companhias Aéreas**: Cria automaticamente companhias se não existirem
3. **Programas de Fidelidade**: Cria automaticamente programas se não existirem
4. **Classes de Serviço**: Converte strings para enum ServiceClassCode
5. **Datas**: Converte strings de data para time.Time
6. **Conexões**: Salva conexões com sequência correta

### Tratamento de Dados

- **Fees**: Suporta tanto estrutura GeminiFees quanto strings
- **Duplicatas**: Usa FirstOrCreate para evitar duplicatas
- **Transações**: Todas as operações são transacionais
- **Validação**: Valida dados obrigatórios antes de salvar

### Mapeamento de Classes de Serviço

```go
"ECONOMY" / "ECONÔMICA" -> ServiceClassEconomy
"PREMIUM ECONOMY" / "ECONOMIA PREMIUM" -> ServiceClassPremiumEconomy
"BUSINESS" / "EXECUTIVA" -> ServiceClassBusiness
"FIRST" / "PRIMEIRA" -> ServiceClassFirst
Outros -> ServiceClassOther
```

## Exemplo de Uso

```go
// Criar serviço
service := NewFlightAwardService(db)

// Dados do Gemini
geminiResp := &schemas.GeminiResponse{
    // ... dados da resposta
}

// Salvar no banco
flightAward, err := service.SaveGeminiResponse(geminiResp)
if err != nil {
    log.Printf("Erro: %v", err)
    return
}

fmt.Printf("Salvo com ID: %s", flightAward.ID)
```

## Tratamento de Erros

O sistema trata os seguintes tipos de erro:

- Dados obrigatórios ausentes
- Formatos de data inválidos
- Erros de banco de dados
- Violações de constraint
- Falhas de transação

Todos os erros são logados e retornados de forma estruturada nas respostas da API.
