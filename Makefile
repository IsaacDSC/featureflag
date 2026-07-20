.DEFAULT_GOAL := help

# Carrega variáveis do .env (se existir) para os targets que sobem a aplicação.
ENV_FILE ?= .env

.PHONY: help run build test tidy fmt vulncheck ci infra-up infra-down docker-up docker-down logs

help: ## Lista os comandos disponíveis
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

run: ## Inicia a aplicação (carrega .env e sobe o server em :3000)
	@set -a; \
	export SECRET_KEY=$${SECRET_KEY:-secret}; \
	export SERVICE_CLIENT_AT=$${SERVICE_CLIENT_AT:-service}; \
	export SDK_CLIENT_AT=$${SDK_CLIENT_AT:-sdk}; \
	[ -f $(ENV_FILE) ] && . ./$(ENV_FILE); \
	set +a; \
	go run ./cmd

build: ## Compila todos os pacotes
	go build -v ./...

test: ## Roda a suíte de testes com o race detector
	go test ./... --race

tidy: ## Ajusta o go.mod/go.sum
	go mod tidy

fmt: ## Verifica se o código está formatado (gofmt)
	@fmt_out="$$(gofmt -l .)"; \
	if [ -n "$$fmt_out" ]; then \
		echo "Arquivos não formatados:"; \
		echo "$$fmt_out"; \
		exit 1; \
	fi

vulncheck: ## Roda o govulncheck (instala se necessário)
	@command -v govulncheck >/dev/null 2>&1 || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

ci: fmt test build vulncheck ## Roda localmente as mesmas validações do CI

infra-up: ## Sobe apenas as dependências (MongoDB)
	docker-compose up -d mongodb

infra-down: ## Derruba as dependências
	docker-compose stop mongodb

docker-up: ## Sobe a stack completa (app + MongoDB + mongo-express)
	docker-compose up -d

docker-down: ## Derruba a stack completa
	docker-compose down

logs: ## Acompanha os logs da stack Docker
	docker-compose logs -f
