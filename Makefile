test:
	cd backend && go test -count=1 ./...

DB_URL=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable

migrate-up:
	migrate -path backend/migrations -database "$(DB_URL)" -verbose up

migrate-down:
	migrate -path backend/migrations -database "$(DB_URL)" -verbose down

migrate-create:
	migrate create -ext sql -dir backend/migrations -seq $(name)

migrate-force:
	migrate -path backend/migrations -database "$(DB_URL)" force $(version)