BIN_DIR      		  := $(CURDIR)/bin
APP_NAME     	      ?= backend
VERSION      		  ?= $(shell git describe --tags 2>/dev/null || echo "v0.0.0")
BUILD_DATE			  := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT   		  := $(shell git rev-parse HEAD 2>/dev/null)
LDFLAGS      		  := "-s -w -X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) -X main.date=$(BUILD_DATE)"
GOLANGCI_LINT_VERSION := v1.62.2
SWAG_VERSION          := v1.16.6
MOCKGEN_VERSION       := v0.4.0
ENV 				  := "local"
UNAME_S 			  := $(shell uname -s)
CONFIG_FILE           ?= .build/config/local.yaml
CONFIG_PATH           := .build/config/
SWAGGER_DOCS_PATH     := internal/application/http_server/swagger/docs

ifeq ($(UNAME_S),Darwin)
SED_INPLACE := -i ''
else
SED_INPLACE := -i
endif

ifeq ($(OS),Windows_NT)
    CONFIG_FILE := $(subst /,\,$(CONFIG_FILE))
endif

.PHONY: all
all: test build

## -- Dependency Management --
.PHONY: install-tools
install-tools: $(BIN_DIR)/golangci-lint $(BIN_DIR)/swag $(BIN_DIR)/mockgen

.PHONY: $(BIN_DIR)/golangci-lint
$(BIN_DIR)/golangci-lint:
	@mkdir -p $(@D)
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@GOBIN=$(BIN_DIR) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: $(BIN_DIR)/swag
$(BIN_DIR)/swag:
	@mkdir -p $(@D)
	@echo "Installing swag $(SWAG_VERSION)..."
	@GOBIN=$(BIN_DIR) go install github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION)

.PHONY: $(BIN_DIR)/mockgen
$(BIN_DIR)/mockgen:
	@mkdir -p $(@D)
	@echo "Installing mockgen $(MOCKGEN_VERSION)..."
	@GOBIN=$(BIN_DIR) go install go.uber.org/mock/mockgen@$(MOCKGEN_VERSION)


## -- Code generating --
.PHONY: generate
generate: generate-mocks docs

## -- Mocks generating --
.PHONY: generate-mocks
generate-mocks: $(BIN_DIR)/mockgen
	@echo "Generating mocks..."

## -- Linting --
.PHONY: lint
lint: $(BIN_DIR)/golangci-lint
	@echo "Running linter..."
	@$(BIN_DIR)/golangci-lint run --config .golangci-lint.yaml --fix=false --color=always

## -- Documentation --
.PHONY: docs
docs: $(BIN_DIR)/swag
	@echo "Generating Swagger documentation..."
	@if [ -n "$(SWAGGER_HOST)" ]; then \
		echo "Using SWAGGER_HOST: $(SWAGGER_HOST)"; \
		$(MAKE) .generate-swagger-with-host HOST=$(SWAGGER_HOST); \
	else \
		echo "Using default host: localhost:8000"; \
		mkdir -p $(SWAGGER_DOCS_PATH); \
		$(BIN_DIR)/swag init -g cmd/api/main.go -o $(SWAGGER_DOCS_PATH) --parseDependency --parseInternal; \
		echo "Swagger documentation generated in $(SWAGGER_DOCS_PATH)"; \
	fi

.PHONY: .generate-swagger-with-host
.generate-swagger-with-host: $(BIN_DIR)/swag
	@echo "Генерация Swagger документации с host: $(HOST)"
	@mkdir -p $(SWAGGER_DOCS_PATH)
	@TEMP_MAIN="cmd/main_temp.go" && \
	cp cmd/api/main.go $$TEMP_MAIN && \
	ESC_HOST=$$(printf '%s\n' "$(HOST)" | sed -e 's/[\\\/&]/\\\\&/g') && \
	sed $(SED_INPLACE) "s|localhost:8000|$${ESC_HOST}|g" $$TEMP_MAIN && \
	$(BIN_DIR)/swag init -g $$TEMP_MAIN -o $(SWAGGER_DOCS_PATH) --parseDependency --parseInternal && \
	rm -f $$TEMP_MAIN && \
	echo "Swagger документация успешно сгенерирована с host: $(HOST)" && \
	echo "Документация доступна по адресу: $(HOST)/swagger/index.html"

## -- Run --
## Local run (non-docker). Override: make run CONFIG_FILE=.build/config/local.yaml
.PHONY: run
run:
	@echo "Starting $(APP_NAME)..."
	@echo CGO_ENABLED=0 go run -trimpath -ldflags=$(LDFLAGS) ./cmd/api/main.go --config_path=$(CONFIG_FILE)
	@CGO_ENABLED=0 go run -trimpath -ldflags=$(LDFLAGS) ./cmd/api/main.go --config_path=$(CONFIG_FILE)

.PHONY: run-notification
run-notification:
	@echo "Starting notification service..."
	@CGO_ENABLED=0 go run -trimpath ./cmd/notification/main.go --config_path=.build/config/notification.yaml

## -- Building --
.PHONY: build
build: build-linux build-darwin build-windows

# Build for Linux
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags=$(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)_linux_amd64 ./cmd/api/main.go
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags=$(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)_linux_arm64 ./cmd/api/main.go

# Build for macOS (Darwin)
.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags=$(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)_darwin_amd64 ./cmd/api/main.go
	@GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags=$(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)_darwin_arm64 ./cmd/api/main.go

# Build for Windows
.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags=$(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)_windows_amd64.exe ./cmd/api/main.go
	@GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags=$(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME)_windows_arm64.exe ./cmd/api/main.go

## -- Docker --
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build \
	    --build-arg BUILD_LDFLAGS=$(LDFLAGS) \
	    --build-arg SWAGGER_HOST=$(SWAGGER_HOST) \
	    --build-arg CONFIG_FILE_PATH="$(CONFIG_PATH)$(ENV).yaml" \
	    -t $(APP_NAME):$(VERSION) \
	    .

.PHONY: compose-up
compose-up:
	@echo "Starting local stack..."
	@docker compose up --build -d

.PHONY: compose-down
compose-down:
	@echo "Stopping local stack..."
	@docker compose down

.PHONY: compose-logs
compose-logs:
	@docker compose logs -f backend notification

.PHONY: docker-build-staging
docker-build-staging:
	@echo "Building Docker image for staging..."
	@docker build \
	    --build-arg BUILD_LDFLAGS=$(LDFLAGS) \
	    --build-arg SWAGGER_HOST=api-staging.example.com \
	    --build-arg CONFIG_FILE_PATH="$(CONFIG_PATH)$(ENV).yaml" \
	    -t $(APP_NAME):staging \
	    .

.PHONY: docker-build-production
docker-build-production:
	@echo "Building Docker image for production..."
	@docker build \
	    --build-arg BUILD_LDFLAGS=$(LDFLAGS) \
	    --build-arg SWAGGER_HOST=api.example.com \
	    --build-arg CONFIG_FILE_PATH="$(CONFIG_PATH)$(ENV).yaml" \
	    -t $(APP_NAME):production \
	    .

## -- Cleanup --
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	@rm -f code-quality-report.json coverage.out coverage.html


## -- Testing --
.PHONY: test
test:
	@echo "Running all unit tests..."
	@go test -v -count=1 ./...
.PHONY: test-services
test-services:
	@echo "Running service tests only..."
	@go test -v -count=1 ./internal/services/...
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -count=1 -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
