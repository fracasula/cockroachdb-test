.PHONY: run
run:
	docker-compose up --force-recreate

.PHONY: sql
sql:
	docker-compose exec roach1 ./cockroach sql --insecure -d="bank"

.PHONY: bootstrap
bootstrap:
	cat bootstrap.sql | docker-compose exec -T roach1 ./cockroach sql --insecure --echo-sql

.PHONY: bootstrap-sqlite
bootstrap-sqlite:
	# sudo apt install sqlite3
	cat bootstrap_sqlite.sql | sqlite3 goapp/sqlite.db < bootstrap_sqlite.sql

.PHONY: test
test: bootstrap
	cd goapp && MODE=CRDB go run main.go

.PHONY: test-sqlite
test-sqlite: bootstrap-sqlite
	cd goapp && MODE=SQLITE go run main.go

.PHONY: rm
rm:
	docker-compose down; docker-compose rm
