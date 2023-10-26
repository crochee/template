OUT_DIR = ./out
PROJECT = template
TARGET_GOARCH=$(ARCH)

# 获取宿主机CPU架构
ifeq ($(TARGET_GOARCH),)
sys=$(shell arch)
ifeq ($(sys),x86_64)
TARGET_GOARCH=amd64
else ifeq ($(sys),AMD64)
TARGET_GOARCH=amd64
else ifeq ($(sys),x64)
TARGET_GOARCH=amd64
else ifeq ($(sys),aarch64)
TARGET_GOARCH=arm64
else ifeq ($(sys),arm64)
TARGET_GOARCH=arm64
else
TARGET_GOARCH=amd64
endif
endif

.PHONY: all
all: clean build
	@mkdir -p $(OUT_DIR)/config && make copy_migration

.PHONY: dep
dep:
	@go mod download && go mod tidy

.PHONY: fmt
fmt:
	@echo "formatting..."
	@gofmt -w .

.PHONY: build
build:
	@make dep &&mkdir -p $(OUT_DIR) &&GOOS=linux CGO_ENABLED=0 GOARCH=$(TARGET_GOARCH) go build -tags jsoniter -o $(OUT_DIR)/$(PROJECT) ./cmd/$(PROJECT) && \
	GOOS=linux CGO_ENABLED=0 GOARCH=$(TARGET_GOARCH) go build -o $(OUT_DIR)/goose ./cmd/goose

.PHONY: copy_migration
copy_migration:
	@mkdir -p $(OUT_DIR) && cp -r internal/store/migration $(OUT_DIR)/

.PHONY: build_goose
build_goose:
	@mkdir -p $(OUT_DIR) && CGO_ENABLED=0 GOARCH=$(TARGET_GOARCH) go build -o $(OUT_DIR)/goose ./cmd/goose

.PHONY: migrate
migrate: build_goose copy_migration
	@$(OUT_DIR)/goose -dir $(OUT_DIR)/migration up

.PHONY: run
run: dep fmt
	@CGO_ENABLED=0 GOARCH=$(TARGET_GOARCH) go run -tags jsoniter ./cmd/$(PROJECT)

.PHONY: clean
clean:
	@rm -rf $(OUT_DIR)

.PHONY: stage
stage:
	@chmod +x ./scripts/stage.sh && ./scripts/stage.sh $(commit)
