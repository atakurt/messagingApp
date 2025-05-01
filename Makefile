APP_NAME = messagingApp
CMD_DIR = ./cmd/api
SWAG_OUTPUT = ./docs
DEV_IMAGE = messaging-app-dev
PROD_IMAGE = messaging-app

.PHONY: build run swag docker-dev docker-prod clean

build:
	go build -o $(APP_NAME) $(CMD_DIR)

run:
	ENV=development ./$(APP_NAME)

swag:
	go install github.com/swaggo/swag/cmd/swag@latest
	$(HOME)/go/bin/swag init --generalInfo $(CMD_DIR)/main.go --output $(SWAG_OUTPUT)

docker-dev:
	docker build -f Dockerfile.test -t $(DEV_IMAGE) .

docker-prod:
	docker build -f Dockerfile -t $(PROD_IMAGE) .

clean:
	rm -f $(APP_NAME)
	rm -rf $(SWAG_OUTPUT)