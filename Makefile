test:
	cd backend && go test -count=1 ./...

PROJECT_ARCHIVE_NAME = Skowronski-Czuba-rozproszony-system-plikow-projekt
zip-project:
	rm -rf $(PROJECT_ARCHIVE_NAME) $(PROJECT_ARCHIVE_NAME).zip
	mkdir -p $(PROJECT_ARCHIVE_NAME)
	rsync -a \
		--exclude='node_modules' \
		--exclude='.git' \
		--exclude='frontend/dist' \
		--exclude='backend/bin' \
		--exclude='backend/data' \
		--exclude='*.log' \
		--exclude='.DS_Store' \
		--exclude='$(PROJECT_ARCHIVE_NAME)' \
		. $(PROJECT_ARCHIVE_NAME)/
	zip -9 -r $(PROJECT_ARCHIVE_NAME).zip $(PROJECT_ARCHIVE_NAME)
	rm -rf $(PROJECT_ARCHIVE_NAME)

DOCS_ARCHIVE_NAME = Skowronski-Czuba-rozproszony-system-plikow-dokumentacja
zip-docs:
	rm -rf $(DOCS_ARCHIVE_NAME) $(DOCS_ARCHIVE_NAME).zip
	mkdir -p $(DOCS_ARCHIVE_NAME)
	cp dokumentacja.tex dokumentacja.pdf $(DOCS_ARCHIVE_NAME)/
	zip -9 -r $(DOCS_ARCHIVE_NAME).zip $(DOCS_ARCHIVE_NAME)
	rm -rf $(DOCS_ARCHIVE_NAME)

DB_URL=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable

migrate-up:
	migrate -path backend/migrations -database "$(DB_URL)" -verbose up

migrate-down:
	migrate -path backend/migrations -database "$(DB_URL)" -verbose down

migrate-create:
	migrate create -ext sql -dir backend/migrations -seq $(name)

migrate-force:
	migrate -path backend/migrations -database "$(DB_URL)" force $(version)

clean-data:
	docker-compose down
	docker volume rm distributed-file-system_master_data || true
	docker volume rm distributed-file-system_storage1_data || true
	docker volume rm distributed-file-system_storage2_data || true
	docker volume rm distributed-file-system_storage3_data || true
# also make sure DB is in sync
