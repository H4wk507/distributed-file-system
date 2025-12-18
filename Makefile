test:
	cd backend && go test -count=1 ./...

zip:
	zip -9 -r Skowronski-Czuba-rozproszony-system-plikow-projekt.zip . \
		-x "node_modules/*" \
		-x ".git/*" \
		-x "frontend/node_modules/*" \
		-x "frontend/dist/*" \
		-x "backend/bin/*" \
		-x "backend/data/*" \
		-x "*.log" \
		-x ".DS_Store"

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