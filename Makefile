docker-build:
	docker build -t signature-server .
	
docker-run:
	docker run --publish 3001:3001 --name signature-server --rm signature-server
	
go-run:
	go run main.go

keygen:
	go run utils/generate.go

vegeta attack:
	vegeta attack -duration=120s -rate=100 -targets=./loadtest/targets.list -output=./loadtest/results/bin/attack.bin \
	&& vegeta plot -title=Vegeta-Attack-Results ./loadtest/results/bin/attack.bin > ./loadtest/results/html/results.html