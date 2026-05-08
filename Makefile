.PHONY: build-linux build-scanner build-frontend build deploy deploy-scanner clean

BINARY=app-manager
SCANNER_BINARY=discovered-app
FRONTEND_DIR=manager-frontend-app
BACKEND_DIR=app-manager
SCANNER_DIR=discovered-app

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -ldflags="-s -w" -o $(BACKEND_DIR)/$(BINARY) ./$(BACKEND_DIR)/main.go

build-scanner:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		go build -ldflags="-s -w" -o $(SCANNER_DIR)/$(SCANNER_BINARY) ./$(SCANNER_DIR)/main.go

build-frontend:
	cd $(FRONTEND_DIR) && npm ci && npm run build

build: build-linux build-scanner build-frontend

deploy: build
	scp $(BACKEND_DIR)/$(BINARY) $(DEPLOY_HOST):/usr/local/bin/
	scp $(BACKEND_DIR)/config.example.yaml $(DEPLOY_HOST):/etc/app-manager/config.yaml
	rsync -avz $(FRONTEND_DIR)/dist/ $(DEPLOY_HOST):/var/www/app-manager/
	ssh $(DEPLOY_HOST) 'systemctl restart app-manager'

deploy-scanner: build-scanner
	scp $(SCANNER_DIR)/$(SCANNER_BINARY) $(DEPLOY_HOST):/usr/local/bin/
	scp scripts/discovered-app.service $(DEPLOY_HOST):/etc/systemd/system/
	ssh $(DEPLOY_HOST) 'systemctl daemon-reload && systemctl enable discovered-app && systemctl restart discovered-app'

clean:
	rm -f $(BACKEND_DIR)/$(BINARY)
	rm -f $(SCANNER_DIR)/$(SCANNER_BINARY)
	rm -rf $(FRONTEND_DIR)/dist $(FRONTEND_DIR)/node_modules
