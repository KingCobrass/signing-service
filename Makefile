docker-build:
	docker build -t signature-service .
	
docker-run:
	docker run --publish 8080:8080 --name signature-service --rm signature-service
	
go-run:
	go run main.go

vegeta attack:
	vegeta attack -duration=120s -rate=100 -targets=./loadtest/targets.list -output=./loadtest/responses/bin/attack.bin \
	&& vegeta plot -title=Vegeta-Attack-Results ./loadtest/responses/bin/attack.bin > ./loadtest/responses/html/results.html