all:
	go test cmd/agent/*.go
	go test cmd/server/*.go
	go build -o cmd/server/server cmd/server/*.go
	go build -o cmd/agent/agent cmd/agent/*.go 
	./metricstest -test.v -test.run=^TestIteration1$ -agent-binary-path=cmd/agent/agent