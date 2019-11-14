include github.com/msales/make/golang

# Run all benchmarks
bench:
	@go test -bench=. $(shell go list ./... | grep -v /vendor/)
.PHONY: bench

# Build the docker image
docker:
	docker build -t double-team --build-arg GITHUB_TOKEN=${GITHUB_TOKEN} .
.PHONY: docker

consume:
	@docker-compose exec -T kafka \
		kafka-console-consumer.sh \
		--bootstrap-server localhost:9092 \
		--from-beginning \
		--topic=test_topic
	@echo Done
.PHONY: consume

DATA=`cat testdata/data.jsonl`
populate:
	@for file in $(DATA); do \
		curl localhost:8082 -X POST -d $${file} ; \
		printf ">" ; \
	done ;
	@echo Done
.PHONY: populate