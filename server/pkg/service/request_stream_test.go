package service

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	gproto "google.golang.org/protobuf/proto"
	"toolplane/pkg/model"
	"toolplane/pkg/trace"
	proto "toolplane/proto"
)

func TestGRPCServerGetRequestChunksReturnsRetainedWindowMetadata(t *testing.T) {
	server, requestService, sessionID := newRequestStreamTestServer(t)

	request, err := requestService.CreateRequest(sessionID, "echo", `{"message":"stream"}`)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	if err := requestService.AppendRequestChunks(sessionID, request.ID, makeStreamChunks(105), model.ResultTypeStreaming); err != nil {
		t.Fatalf("append chunks: %v", err)
	}

	response, err := server.GetRequestChunks(context.Background(), &proto.GetRequestChunksRequest{
		SessionId: sessionID,
		RequestId: request.ID,
	})
	if err != nil {
		t.Fatalf("get request chunks: %v", err)
	}

	if response.GetStartSeq() != 6 {
		t.Fatalf("start_seq = %d, want 6", response.GetStartSeq())
	}
	if response.GetNextSeq() != 106 {
		t.Fatalf("next_seq = %d, want 106", response.GetNextSeq())
	}
	if len(response.GetChunks()) != 100 {
		t.Fatalf("chunk count = %d, want 100", len(response.GetChunks()))
	}
	if response.GetChunks()[0] != "chunk-006" {
		t.Fatalf("first retained chunk = %q, want chunk-006", response.GetChunks()[0])
	}
	if response.GetChunks()[99] != "chunk-105" {
		t.Fatalf("last retained chunk = %q, want chunk-105", response.GetChunks()[99])
	}

	again, err := server.GetRequestChunks(context.Background(), &proto.GetRequestChunksRequest{
		SessionId: sessionID,
		RequestId: request.ID,
	})
	if err != nil {
		t.Fatalf("get request chunks again: %v", err)
	}
	if again.GetStartSeq() != response.GetStartSeq() || again.GetNextSeq() != response.GetNextSeq() {
		t.Fatalf("second retained window = [%d,%d), want [%d,%d)", again.GetStartSeq(), again.GetNextSeq(), response.GetStartSeq(), response.GetNextSeq())
	}
}

func TestGRPCServerResumeStreamReplaysRetainedWindowAndFinalMarker(t *testing.T) {
	server, requestService, sessionID := newRequestStreamTestServer(t)

	request, err := requestService.CreateRequest(sessionID, "echo", `{"message":"resume"}`)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	if err := requestService.AppendRequestChunks(sessionID, request.ID, []string{"alpha", "beta"}, model.ResultTypeStreaming); err != nil {
		t.Fatalf("append chunks: %v", err)
	}
	if err := requestService.SubmitRequestResult(sessionID, request.ID, map[string]string{"done": "ok"}, model.ResultTypeResolution, nil); err != nil {
		t.Fatalf("submit result: %v", err)
	}

	stream := newCollectingExecuteToolStream(context.Background())
	err = server.ResumeStream(&proto.ResumeStreamRequest{RequestId: request.ID, LastSeq: 1}, stream)
	if err != nil {
		t.Fatalf("resume stream: %v", err)
	}

	if len(stream.chunks) != 2 {
		t.Fatalf("sent chunk count = %d, want 2", len(stream.chunks))
	}
	if stream.chunks[0].GetSeq() != 2 || stream.chunks[0].GetChunk() != "beta" || stream.chunks[0].GetIsFinal() {
		t.Fatalf("first resumed chunk = %+v, want seq=2 chunk=beta final=false", stream.chunks[0])
	}
	if stream.chunks[1].GetSeq() != 3 || !stream.chunks[1].GetIsFinal() {
		t.Fatalf("final resumed chunk = %+v, want seq=3 final=true", stream.chunks[1])
	}
	if !strings.Contains(stream.chunks[1].GetChunk(), `"done":"ok"`) {
		t.Fatalf("final resumed payload = %q, want JSON result containing done=ok", stream.chunks[1].GetChunk())
	}
}

func TestGRPCServerResumeStreamReturnsOutOfRangeWhenRetainedWindowExpired(t *testing.T) {
	server, requestService, sessionID := newRequestStreamTestServer(t)

	request, err := requestService.CreateRequest(sessionID, "echo", `{"message":"expired"}`)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	if err := requestService.AppendRequestChunks(sessionID, request.ID, makeStreamChunks(105), model.ResultTypeStreaming); err != nil {
		t.Fatalf("append chunks: %v", err)
	}
	if err := requestService.SubmitRequestResult(sessionID, request.ID, map[string]string{"done": "late"}, model.ResultTypeResolution, nil); err != nil {
		t.Fatalf("submit result: %v", err)
	}

	stream := newCollectingExecuteToolStream(context.Background())
	err = server.ResumeStream(&proto.ResumeStreamRequest{RequestId: request.ID, LastSeq: 0}, stream)
	if status.Code(err) != codes.OutOfRange {
		t.Fatalf("resume status = %v, want %v (err=%v)", status.Code(err), codes.OutOfRange, err)
	}
	if len(stream.chunks) != 0 {
		t.Fatalf("sent chunk count = %d, want 0", len(stream.chunks))
	}
}

func TestGRPCServerResumeStreamReplaysTrimmedRetainedWindowAndFinalMarker(t *testing.T) {
	server, requestService, sessionID := newRequestStreamTestServer(t)

	request, err := requestService.CreateRequest(sessionID, "echo", `{"message":"trimmed"}`)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	if err := requestService.AppendRequestChunks(sessionID, request.ID, makeStreamChunks(105), model.ResultTypeStreaming); err != nil {
		t.Fatalf("append chunks: %v", err)
	}
	if err := requestService.SubmitRequestResult(sessionID, request.ID, map[string]string{"done": "trimmed"}, model.ResultTypeResolution, nil); err != nil {
		t.Fatalf("submit result: %v", err)
	}

	stream := newCollectingExecuteToolStream(context.Background())
	err = server.ResumeStream(&proto.ResumeStreamRequest{RequestId: request.ID, LastSeq: 100}, stream)
	if err != nil {
		t.Fatalf("resume stream: %v", err)
	}

	if len(stream.chunks) != 6 {
		t.Fatalf("sent chunk count = %d, want 6", len(stream.chunks))
	}

	for index, expectedChunk := range []string{"chunk-101", "chunk-102", "chunk-103", "chunk-104", "chunk-105"} {
		chunk := stream.chunks[index]
		wantSeq := int32(101 + index)
		if chunk.GetSeq() != wantSeq || chunk.GetChunk() != expectedChunk || chunk.GetIsFinal() {
			t.Fatalf("resumed chunk %d = %+v, want seq=%d chunk=%s final=false", index, chunk, wantSeq, expectedChunk)
		}
	}

	finalChunk := stream.chunks[5]
	if finalChunk.GetSeq() != 106 || !finalChunk.GetIsFinal() {
		t.Fatalf("final resumed chunk = %+v, want seq=106 final=true", finalChunk)
	}
	if !strings.Contains(finalChunk.GetChunk(), `"done":"trimmed"`) {
		t.Fatalf("final resumed payload = %q, want JSON result containing done=trimmed", finalChunk.GetChunk())
	}
}

func newRequestStreamTestServer(t *testing.T) (*GRPCServer, *RequestsService, string) {
	t.Helper()

	toolService := NewToolService(trace.NopTracer(), nil)
	machineService := NewMachinesService(toolService, trace.NopTracer(), nil)
	requestService := NewRequestsService(toolService, machineService, trace.NopTracer(), nil)
	server := NewGRPCServer(toolService, nil, machineService, requestService, nil)

	const sessionID = "session-stream"
	const machineID = "machine-stream"

	_, err := machineService.RegisterMachine(sessionID, machineID, "1.0.0", "go", "127.0.0.1", []*model.Tool{
		model.NewTool(sessionID, machineID, "echo", "echo tool", `{"type":"object"}`, nil, nil),
	})
	if err != nil {
		t.Fatalf("register machine: %v", err)
	}

	return server, requestService, sessionID
}

func makeStreamChunks(count int) []string {
	chunks := make([]string, 0, count)
	for index := 1; index <= count; index++ {
		chunks = append(chunks, fmt.Sprintf("chunk-%03d", index))
	}
	return chunks
}

type collectingExecuteToolStream struct {
	ctx    context.Context
	chunks []*proto.ExecuteToolChunk
}

func newCollectingExecuteToolStream(ctx context.Context) *collectingExecuteToolStream {
	if ctx == nil {
		ctx = context.Background()
	}
	return &collectingExecuteToolStream{ctx: ctx}
}

func (s *collectingExecuteToolStream) Send(chunk *proto.ExecuteToolChunk) error {
	clone, ok := gproto.Clone(chunk).(*proto.ExecuteToolChunk)
	if !ok {
		return fmt.Errorf("unexpected chunk clone type %T", chunk)
	}
	s.chunks = append(s.chunks, clone)
	return nil
}

func (s *collectingExecuteToolStream) SetHeader(metadata.MD) error {
	return nil
}

func (s *collectingExecuteToolStream) SendHeader(metadata.MD) error {
	return nil
}

func (s *collectingExecuteToolStream) SetTrailer(metadata.MD) {}

func (s *collectingExecuteToolStream) Context() context.Context {
	return s.ctx
}

func (s *collectingExecuteToolStream) SendMsg(message interface{}) error {
	chunk, ok := message.(*proto.ExecuteToolChunk)
	if !ok {
		return fmt.Errorf("unexpected message type %T", message)
	}
	return s.Send(chunk)
}

func (s *collectingExecuteToolStream) RecvMsg(interface{}) error {
	return nil
}
