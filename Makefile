.PHONY: build-linux build-frontend build deploy clean

BINARY=app-manager
FRONTEND_DIR=manager-frontend-app
BACKEND_DIR=app-manager

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -ldflags="-s -w" -o $(BACKEND_DIR)/$(BINARY) ./$(BACKEND_DIR)/main.go

build-frontend:
	cd $(FRONTEND_DIR) && npm ci && npm run build

build: build-linux build-frontend

deploy: build
	scp $(BACKEND_DIR)/$(BINARY) $(DEPLOY_HOST):/usr/local/bin/
	scp $(BACKEND_DIR)/config.example.yaml $(DEPLOY_HOST):/etc/app-manager/config.yaml
	rsync -avz $(FRONTEND_DIR)/dist/ $(DEPLOY_HOST):/var/www/app-manager/
	ssh $(DEPLOY_HOST) 'systemctl restart app-manager'

clean:
	rm -f $(BACKEND_DIR)/$(BINARY)
	rm -rf $(FRONTEND_DIR)/dist $(FRONTEND_DIR)/node_modules
