# Main Makefile for the Go server
.PHONY: all build clean install run docker-build docker-run gen-proto gen-proto-python gen-client-protos gen-proto-all check-proto-drift conformance-python conformance-python-grpc conformance-python-http release-gate release-gate-runtime

PROTO_DRIFT_FILES = \
	proto/service.pb.go \
	proto/service_grpc.pb.go \
	proto/service.pb.gw.go \
	../clients/go-client/proto/service.proto \
	../clients/go-client/proto/service.pb.go \
	../clients/go-client/proto/service_grpc.pb.go \
	../clients/go-client/google/api/annotations.proto \
	../clients/go-client/google/api/http.proto \
	../clients/python-client/toolplane/proto/service_pb2.py \
	../clients/python-client/toolplane/proto/service_pb2_grpc.py \
	../clients/typescript-client/proto/service.proto \
	../clients/typescript-client/google/api/annotations.proto \
	../clients/typescript-client/google/api/http.proto \
	../clients/typescript-client/src/proto/proto/service_pb.js \
	../clients/typescript-client/src/proto/proto/service_pb.d.ts \
	../clients/typescript-client/src/proto/proto/service_grpc_pb.js \
	../clients/typescript-client/src/proto/proto/service_grpc_pb.d.ts \
	../clients/typescript-client/src/proto/google/api/annotations_pb.js \
	../clients/typescript-client/src/proto/google/api/annotations_pb.d.ts \
	../clients/typescript-client/src/proto/google/api/annotations_grpc_pb.js \
	../clients/typescript-client/src/proto/google/api/http_pb.js \
	../clients/typescript-client/src/proto/google/api/http_pb.d.ts \
	../clients/typescript-client/src/proto/google/api/http_grpc_pb.js

all: build

TOOLPLANE_ENV_MODE ?= development
TOOLPLANE_AUTH_MODE ?= fixed
TOOLPLANE_AUTH_FIXED_API_KEY ?= toolplane-conformance-fixture-key
TOOLPLANE_CONFORMANCE_API_KEY ?= $(TOOLPLANE_AUTH_FIXED_API_KEY)
TOOLPLANE_STORAGE_MODE ?= $(if $(strip $(TOOLPLANE_DATABASE_URL)),postgres,memory)
TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND ?= 1

# Go server commands
build:
	go build -o bin/toolplane-server ./cmd/server

run:
	./bin/toolplane-server

# Clean commands
clean:
	rm -rf bin

install:
	go mod download

# Docker commands
docker-build:
	docker build -t toolplane-server:latest .

docker-run:
	docker run -p 50051:50051 -p 9090:9090 toolplane-server:latest

# Generate protobuf files
gen-proto:
	# Generate Go types and gRPC & HTTP gateway code from service.proto
	protoc -I. -Iproto -Iproto/googleapis \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative,logtostderr=true \
		proto/service.proto

# Generate protobuf files
gen-proto-python:
	# Generate tracked Python client outputs from service.proto
	python3 -m grpc_tools.protoc \
		--proto_path=. \
		--python_out=../clients/python-client/toolplane \
		--grpc_python_out=../clients/python-client/toolplane \
		proto/service.proto
	python3 scripts/fix_python_proto_imports.py ../clients/python-client/toolplane/proto/service_pb2_grpc.py

gen-client-protos:
	bash ./generate_client_protos.sh

# Canonical full-regeneration entrypoint.
# server/proto/service.proto is the only editable contract source;
# all client-side .proto copies and generated stubs flow from it.
gen-proto-all: gen-proto gen-client-protos

check-proto-drift: gen-proto-all
	git --no-pager diff --exit-code -- $(PROTO_DRIFT_FILES)

# Conformance tests (Python SDK)
conformance-python:
	python -m pytest ../clients/python-client/tests/conformance/test_conformance_runner.py -m conformance -q

conformance-python-grpc:
	python -m pytest ../clients/python-client/tests/conformance/test_conformance_runner.py -m conformance -k grpc -q

conformance-python-http:
	python -m pytest ../clients/python-client/tests/conformance/test_conformance_runner.py -m conformance -k http -q

# Authoritative release gate: proves the maintained architecture end to end.
# See server/docs/release-gate.md for the scenario description.
release-gate:
	TOOLPLANE_ENV_MODE=$(TOOLPLANE_ENV_MODE) \
	TOOLPLANE_AUTH_MODE=$(TOOLPLANE_AUTH_MODE) \
	TOOLPLANE_AUTH_FIXED_API_KEY=$(TOOLPLANE_AUTH_FIXED_API_KEY) \
	TOOLPLANE_CONFORMANCE_API_KEY=$(TOOLPLANE_CONFORMANCE_API_KEY) \
	TOOLPLANE_STORAGE_MODE=$(TOOLPLANE_STORAGE_MODE) \
	TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND=$(TOOLPLANE_PROXY_ALLOW_INSECURE_BACKEND) \
	python -m pytest ../clients/python-client/tests/conformance/test_conformance_runner.py -m conformance -q
	$(MAKE) release-gate-runtime

release-gate-runtime:
	go test ./cmd/server -run 'TestLoadServerConfigRejectsProductionInMemoryStorage|TestLoadServerConfigAllowsProductionPostgresStorage' -count=1
	go test ./pkg/service -run 'TestSessionsServiceRecordsAuditEvents|TestTasksServiceCancelTaskCancelsUnderlyingRequest|TestTasksServiceTimeoutCancelsUnderlyingRequest' -count=1
	go test ./pkg/service -run 'TestRequestsServiceInMemoryLeaseExpiryRequeuesRunningRequest|TestRequestsServicePersistentRecoveryRequeuesExpiredRequest|TestMachinesServiceDrainMachineWaitsForInflightRequestAndBlocksNewWork|TestMachinesServiceDrainMachineWaitsForClaimedRequestUntilLeaseExpiryRequeuesIt|TestGRPCServerResumeStreamReplaysRetainedWindowAndFinalMarker|TestGRPCServerResumeStreamReplaysTrimmedRetainedWindowAndFinalMarker|TestGRPCServerResumeStreamReturnsOutOfRangeWhenRetainedWindowExpired' -count=1