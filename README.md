# Feature Flag

_Simples feature flag / feature toggle_

![Green Teal Geometric Modern Computer Programmer Code Editor Quotes Instagram Post](https://github.com/IsaacDSC/featureflag/assets/56350331/eeb6227d-5a70-4a00-af21-e43368453c60)

## 🎯 Visão Geral

Este projeto oferece dois serviços principais com diferentes garantias do [teorema CAP](https://en.wikipedia.org/wiki/CAP_theorem):

| Serviço | Modelo CAP | Descrição |
|---------|------------|-----------|
| **Feature Flag** | **AP** (Availability + Partition Tolerance) | Prioriza disponibilidade e tolerância a partições. O SDK mantém cache local, garantindo que a aplicação sempre tenha uma resposta, mesmo em caso de falha de rede. Eventual consistency via SSE. |
| **Content Hub** | **CP** (Consistency + Partition Tolerance) | Prioriza consistência e tolerância a partições. Garante que o conteúdo retornado seja sempre o mais atualizado, mesmo que isso signifique maior latência em casos de partição. |

> 📐 **Arquitetura:** Para detalhes sobre a infraestrutura e fluxo de dados, consulte **[docs/ARCH.md](docs/ARCH.md)**

---

## 🚀 Startup do Serviço

### 1. Configurar variáveis de ambiente

Copie o arquivo de exemplo `.env_example` para `.env`:

```sh
cp .env_example .env
```

O arquivo `.env` contém as seguintes configurações:

```sh
export SDK_CLIENT_AT="secret"
export MONGODB_URI="mongodb://admin:password@localhost:27017"
export MONGODB_NAME="featureflag"
export MONGODB_IDX_TIMEOUT="1s"
export REPOSITORY_TYPE="mongodb"
```

> 💡 **Dica:** Se você utiliza [direnv](https://direnv.net/), basta copiar o conteúdo para o arquivo `.envrc` e executar `direnv allow`.

### 2. Iniciar o serviço com Docker

```sh
docker-compose up -d
```

O serviço estará disponível em `http://localhost:3000`.

---

## 📦 Instalação do SDK

```sh
go get -u github.com/IsaacDSC/featureflag
```

---

## 📚 Documentação

### Feature Flag

Documentação completa sobre como criar e utilizar Feature Flags:

👉 **[docs/FEATURE_FLAG.md](docs/FEATURE_FLAG.md)**

- Feature Flag tipo disjuntor (on/off simples)
- Feature Flag com porcentagem (A/B testing)
- Feature Flag com session_id (rollout controlado)
- Exemplos de uso com o SDK Go

### Content Hub

Documentação completa sobre como criar e utilizar o Content Hub:

👉 **[docs/CONTENT_HUB.md](docs/CONTENT_HUB.md)**

- Criação de conteúdo dinâmico
- Estratégias de sessão (personalização por usuário)
- Estratégias de balanceamento (distribuição ponderada)
- Exemplos de uso com o SDK Go

---

## 🔐 Autenticação

### Obter token do SDK Client

```http
POST http://localhost:3000/auth
Authorization: <token>
```

### Obter token do Service Client

```http
POST http://localhost:3000/auth
Authorization: <token>
```

