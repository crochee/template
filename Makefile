OUT_DIR = ./out
PROJECT = go_template

.PHONY: all
all: clean build
	@mkdir -p $(OUT_DIR)/config && make copy_migration

.PHONY: dep
dep:
	@go mod download && go mod tidy

.PHONY: fmt
fmt:
	@echo "formatting..."
	@gofmt -w ./../$(PROJECT)

.PHONY: build
build:
	@make dep &&mkdir -p $(OUT_DIR) &&GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -tags jsoniter -o $(OUT_DIR)/$(PROJECT) ./cmd/$(PROJECT) && \
	GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -o $(OUT_DIR)/goose ./cmd/goose

.PHONY: copy_migration
copy_migration:
	@mkdir -p $(OUT_DIR) && cp -r internal/store/migration $(OUT_DIR)/

.PHONY: build_goose
build_goose:
	@mkdir -p $(OUT_DIR) && CGO_ENABLED=0 GOARCH=amd64 go build -o $(OUT_DIR)/goose ./cmd/goose

.PHONY: migrate
migrate: build_goose copy_migration
	@$(OUT_DIR)/goose -dir $(OUT_DIR)/migration up

.PHONY: run
run: dep fmt
	@CGO_ENABLED=0 GOARCH=amd64 go run -tags jsoniter ./cmd/$(PROJECT)

.PHONY: clean
clean:
	@rm -rf $(OUT_DIR)
