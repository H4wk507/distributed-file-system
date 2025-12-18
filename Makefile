test:
	cd backend && go test -count=1 ./...

ARCHIVE_NAME = Skowronski-Czuba-rozproszony-system-plikow-projekt
zip:
	rm -rf $(ARCHIVE_NAME) $(ARCHIVE_NAME).zip
	mkdir -p $(ARCHIVE_NAME)
	rsync -a \
		--exclude='node_modules' \
		--exclude='.git' \
		--exclude='frontend/dist' \
		--exclude='backend/bin' \
		--exclude='backend/data' \
		--exclude='*.log' \
		--exclude='.DS_Store' \
		--exclude='$(ARCHIVE_NAME)' \
		. $(ARCHIVE_NAME)/
	zip -9 -r $(ARCHIVE_NAME).zip $(ARCHIVE_NAME)
	rm -rf $(ARCHIVE_NAME)

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
