# 🔍 Full-Text Search API com Go

Uma aplicação Go que implementa busca textual usando tanto PostgreSQL quanto Elasticsearch, com monitoramento completo via Elastic APM e Kibana.

## 📋 Índice

- [Visão Geral](#-visão-geral)
- [Arquitetura](#-arquitetura)
- [Funcionalidades](#-funcionalidades)
- [Pré-requisitos](#-pré-requisitos)
- [Configuração e Execução](#-configuração-e-execução)
- [Endpoints da API](#-endpoints-da-api)
- [Comparação: PostgreSQL vs Elasticsearch](#-comparação-postgresql-vs-elasticsearch)
- [Monitoramento com APM](#-monitoramento-com-apm)
- [Estrutura do Projeto](#-estrutura-do-projeto)
- [Casos de Uso](#-casos-de-uso)

## 🎯 Visão Geral

Este projeto demonstra a implementação de dois tipos de busca textual:

1. **Busca Tradicional (PostgreSQL)**: Utiliza `ILIKE` para busca por padrões
2. **Busca Otimizada (Elasticsearch)**: Implementa busca textual completa com relevância

A aplicação sincroniza automaticamente dados do PostgreSQL para o Elasticsearch e oferece observabilidade completa através do Elastic APM.

## 🏗️ Arquitetura

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Cliente   │───►│  API Go     │───►│ PostgreSQL  │
└─────────────┘    │             │    └─────────────┘
                   │             │           │
                   │             │           │ Sincronização
                   │             │           ▼
                   │             │    ┌─────────────┐
                   │             │───►│Elasticsearch│
                   └─────────────┘    └─────────────┘
                          │                   │
                          ▼                   │
                   ┌─────────────┐           │
                   │ APM Server  │◄──────────┘
                   └─────────────┘
                          │
                          ▼
                   ┌─────────────┐
                   │   Kibana    │
                   └─────────────┘
```

## ✨ Funcionalidades

- **🔍 Busca Dupla**: PostgreSQL (padrões) + Elasticsearch (relevância)
- **🔄 Sincronização Automática**: Dados do PostgreSQL são indexados no Elasticsearch
- **📊 Observabilidade Completa**: APM para monitoramento de performance
- **🐳 Containerização**: Docker Compose para fácil deployment
- **⚡ Comparação de Performance**: Compare tempos de resposta entre os backends
- **📈 Dashboards**: Interface visual no Kibana para análise

## 🛠️ Pré-requisitos

- **Docker** e **Docker Compose**
- **Go 1.23+** (para desenvolvimento local)

## 🚀 Configuração e Execução

### 1. Clone e Navegue

```bash
git clone <seu-repositorio>
cd full-text-search
```

### 2. Inicie os Serviços

```bash
# Inicia todos os serviços (PostgreSQL, Elasticsearch, Kibana, APM)
docker-compose up -d

# Acompanhe os logs (opcional)
docker-compose logs -f
```

### 3. Aguarde a Inicialização

- **PostgreSQL**: ~ 1-2 minutos (incluindo inserção de 500k registros)
- **Elasticsearch**: ~1-2 minutos
- **Sincronização de Dados**: ~3-5 minutos
- **Kibana**: ~2-3 minutos

### 4. Verifique os Serviços

```bash
# PostgreSQL
curl http://localhost:8081/health

# Elasticsearch 
curl http://localhost:9200/_cluster/health

# Kibana
curl http://localhost:5601/api/status
```

## 🔗 Endpoints da API

### Base URL: `http://localhost:8081`

### 1. **Busca PostgreSQL**
```bash
GET /search?query=<termo>

# Exemplo
curl "http://localhost:8081/search?query=sample_text_74_elderberry_74"
```

### 2. **Busca Elasticsearch (Otimizada)**
```bash
GET /search/optimized?query=<termo>

# Exemplo  
curl "http://localhost:8081/search/optimized?query=sample_text_74_elderberry_74"
```

### 3. **Health Check**
```bash
GET /health

# Exemplo
curl "http://localhost:8081/health"
```

### 📄 Formato de Resposta

```json
{
  "results": [
    {
      "id": 74,
      "text": "sample_text_74_elderberry_74"
    }
  ],
  "query": "sample_text_74_elderberry_74"
}
```

## ⚖️ Comparação: PostgreSQL vs Elasticsearch

### PostgreSQL (`/search`)

**✅ Vantagens:**
- **Simplicidade**: Configuração mínima
- **Consistência**: Dados sempre atualizados
- **Familiaridade**: SQL padrão
- **Menos Recursos**: Menor uso de memória
- **Atomicidade**: Transações ACID

**❌ Desvantagens:**
- **Performance**: Lenta em grandes volumes (>1M registros)
- **Funcionalidades Limitadas**: Apenas busca por padrões (`ILIKE`)
- **Escalabilidade**: Não escala horizontalmente para buscas
- **Relevância**: Sem ranking de resultados

### Elasticsearch (`/search/optimized`)

**✅ Vantagens:**
- **Performance Superior**: 10-100x mais rápida em grandes datasets
- **Busca Inteligente**: Análise de texto, stemming, synonyms
- **Relevância**: Score de relevância para ranking
- **Escalabilidade**: Escala horizontalmente
- **Flexibilidade**: Queries complexas, filtros, agregações
- **Análise de Texto**: Suporte a múltiplas linguagens

**❌ Desvantagens:**
- **Complexidade**: Configuração e manutenção mais complexas
- **Recursos**: Alto uso de memória (mín. 2GB)
- **Consistência Eventual**: Pequeno delay na sincronização
- **Dependência Adicional**: Mais um serviço para manter

## 📊 Monitoramento com APM

### Acesse o Kibana

1. **URL**: http://localhost:5601
2. **Navegue**: Observability → APM → Services
3. **Selecione**: `full-text-search-api`

### 🔍 Métricas Disponíveis

- **Tempo de Resposta**: Compare `/search` vs `/search/optimized`
- **Throughput**: Requisições por minuto
- **Erro Rate**: Percentual de erros
- **Database Queries**: Performance das consultas SQL
- **Elasticsearch Traces**: Tempo de busca no ES
- **Service Map**: Visualização das dependências

### 📈 Dashboards Principais

1. **Service Overview**: Métricas gerais da aplicação
2. **Transactions**: Performance por endpoint
3. **Dependencies**: Mapa de serviços
4. **Errors**: Análise de erros

## 📁 Estrutura do Projeto

```
├── main.go              # Aplicação principal
├── docker-compose.yml   # Orquestração dos serviços
├── Dockerfile          # Build da aplicação Go
├── init.sql           # Schema e dados iniciais
├── go.mod             # Dependências Go
├── go.sum             # Lock das dependências
└── README.md          # Esta documentação
```

### 🔧 Componentes Principais

**`main.go`**:
- Handler para endpoints `/search` e `/search/optimized`
- Instrumentação APM automática
- Sincronização PostgreSQL → Elasticsearch

**`docker-compose.yml`**:
- PostgreSQL (porta 5432)
- Elasticsearch (porta 9200)  
- Kibana (porta 5601)
- APM Server (porta 8200)
- API Go (porta 8081)

**`init.sql`**:
- Criação da tabela `search_data`
- Inserção de 500.000 registros de teste
- Dados variados para demonstrar diferentes padrões de busca

## 🎯 Casos de Uso

### 1. **Aplicações Pequenas/Médias** (< 100k registros)
```
Recomendação: PostgreSQL
Motivo: Simplicidade, menor overhead, consistência
```

### 2. **E-commerce/Catálogos** (> 100k produtos)
```
Recomendação: Elasticsearch
Motivo: Busca por relevância, facetas, performance
```

### 3. **Logs/Analytics** (> 1M registros)
```
Recomendação: Elasticsearch
Motivo: Agregações, time-series, escalabilidade
```

### 4. **Busca Empresarial** (documentos complexos)
```
Recomendação: Elasticsearch
Motivo: Full-text search, múltiplos campos, highlighting
```

## 🔧 Desenvolvimento Local

### Executar sem Docker

```bash
# 1. Inicie apenas PostgreSQL e Elasticsearch
docker-compose up -d postgres elasticsearch

# 2. Configure variáveis de ambiente
export DB_HOST=localhost
export DB_PORT=5435
export DB_USER=postgres  
export DB_PASSWORD=password
export DB_NAME=searchdb
export ELASTICSEARCH_URL=http://localhost:9200

# 3. Execute a aplicação
go run main.go
```

### Testes de Performance

```bash
# Teste PostgreSQL (1000 requisições)
ab -n 1000 -c 10 "http://localhost:8081/search?query=apple"

# Teste Elasticsearch (1000 requisições)  
ab -n 1000 -c 10 "http://localhost:8081/search/optimized?query=apple"
```

## 🧹 Limpeza

```bash
# Para todos os serviços
docker-compose down

# Remove volumes (deleta dados)
docker-compose down -v

# Remove imagens
docker-compose down --rmi all -v
```

## 📝 Logs de Debug

```bash
# API Logs
docker-compose logs -f api

# Elasticsearch Logs
docker-compose logs -f elasticsearch

# APM Server Logs
docker-compose logs -f apm-server
```

## 🚀 Deploy em Produção

### Considerações Importantes:

1. **Recursos**: Elasticsearch precisa de pelo menos 4GB RAM
2. **Segurança**: Habilitar autenticação (removemos xpack.security para simplificar)
3. **Backup**: Configurar snapshots do Elasticsearch
4. **Monitoring**: Usar Elastic Cloud ou configurar cluster monitoring
5. **Performance**: Configurar number_of_shards baseado no volume de dados

---

## 🤝 Contribuição

Sinta-se à vontade para abrir issues e pull requests para melhorar este projeto!

## 📄 Licença

MIT License - veja o arquivo LICENSE para detalhes.