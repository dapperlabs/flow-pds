.PHONY: dev
dev:
	docker-compose up -d db pgadmin emulator
	docker-compose logs -f

.PHONY: stop
stop:
	docker-compose stop

.PHONY: down
down:
	docker-compose down

.PHONY: reset
reset: down dev

.PHONY: test
test:
	go test ./...

.PHONY: test-clean
test-clean:
	go clean -testcache && go test ./...

.PHONY: deploy
deploy:
	flow project deploy --update
