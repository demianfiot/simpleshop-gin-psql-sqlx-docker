.PHONY: migrate-up migrate-down migrate-create klickhouse-migrate-up db-shell


include .env
export

# Міграції  /schema
migrate-up:
	docker run --rm \
		-v $(PWD)/schema:/schema \
		--network gptprac3_crm-network \
		migrate/migrate \
		-path=/schema \
		-database 'postgres://$(DB_USER):$(DB_PASSWORD)@postgres:5432/$(DB_NAME)?sslmode=disable' \
		up

migrate-down:
	docker run --rm \
		-v $(PWD)/schema:/schema \
		--network gptprac3_crm-network \
		migrate/migrate \
		-path=/schema \
		-database 'postgres://$(DB_USER):$(DB_PASSWORD)@postgres:5432/$(DB_NAME)?sslmode=disable' \
		down 1

migrate-create:
	@read -p "Migration name: " name; \
	mkdir -p schema; \
	docker run --rm \
		-v $(PWD)/schema:/schema \
		migrate/migrate \
		create -ext sql -dir /schema -seq $${name}
		
klickhouse-migrate-up:
	docker exec -it clickhouse clickhouse-client \
	--query "$(cat schema/clickhouse/001_create_orders_analytics.sql)"

klickhousedb:
	docker exec -it clickhouse clickhouse-client \
  --user $(CLICKHOUSE_USER) \
  --password $(CLICKHOUSE_PASSWORD) \
  --database $(CLICKHOUSE_DATABASE)
# команди
up:
	docker-compose up -d

down:
	docker-compose down -v 

logs:
	docker-compose logs -f backend

db-shell:
	docker-compose exec postgres psql -U $(DB_USER) -d $(DB_NAME)

status:
	docker-compose ps
	docker-compose exec postgres psql -U $(DB_USER) -d $(DB_NAME) -c "\dt"

