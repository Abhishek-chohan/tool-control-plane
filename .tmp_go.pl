my $n = 0;
sub edit {
  my ($file, $old, $new) = @_;
  open(my $f, '<', $file) or die "$file: $!"; local $/; my $src = <$f>; close($f);
  die "missing in $file:\n$old\n" unless index($src, $old) >= 0;
  $src =~ s/\Q$old\E/$new/;
  open(my $o, '>', $file) or die $!; print $o $src; close($o);
  $n++;
}

my $cf = 'clients/go-client/client/toolplane_client.go';

# 1. executionTimeout field, defaulted; honored in executeToolGRPC,
#    StreamExecuteTool, waitForRequestCompletion.
edit($cf,
  "\tuserID     string\n\tapiKey     string\n\ttlsConfig  GRPCTLSConfig\n",
  "\tuserID     string\n\tapiKey     string\n\ttlsConfig  GRPCTLSConfig\n\t// executionTimeout bounds ExecuteTool waits and execution streams.\n\t// Defaults to defaultGRPCExecutionTimeout; override with\n\t// WithExecutionTimeout.\n\texecutionTimeout time.Duration\n");

# 2. default it in the constructor.
edit($cf,
  "\tclient := &ToolplaneClient{\n\t\tprotocol:   protocol,\n\t\tserverHost: serverHost,\n\t\tserverPort: serverPort,\n\t\tsessionID:  sessionID,\n\t\tuserID:     userID,\n\t\tapiKey:     apiKey,\n\t}",
  "\tclient := &ToolplaneClient{\n\t\tprotocol:         protocol,\n\t\tserverHost:       serverHost,\n\t\tserverPort:       serverPort,\n\t\tsessionID:        sessionID,\n\t\tuserID:           userID,\n\t\tapiKey:           apiKey,\n\t\texecutionTimeout: defaultGRPCExecutionTimeout,\n\t}");

# 3. honor it in executeToolGRPC.
edit($cf,
  "\texecCtx, cancel := c.grpcContext(ctx, defaultGRPCExecutionTimeout)\n\tdefer cancel()\n\n\tresponse, err := c.toolClient.ExecuteTool(execCtx, request)",
  "\texecCtx, cancel := c.executionContext(ctx)\n\tdefer cancel()\n\n\tresponse, err := c.toolClient.ExecuteTool(execCtx, request)");

# 4. honor it in StreamExecuteTool.
edit($cf,
  "\tstreamCtx, cancel := c.grpcContext(ctx, defaultGRPCExecutionTimeout)\n\tdefer cancel()\n\n\tstream, err := c.toolClient.StreamExecuteTool(streamCtx, &pb.ExecuteToolRequest{",
  "\tstreamCtx, cancel := c.executionContext(ctx)\n\tdefer cancel()\n\n\tstream, err := c.toolClient.StreamExecuteTool(streamCtx, &pb.ExecuteToolRequest{");

# 5. honor it in waitForRequestCompletion.
edit($cf,
  "\texecCtx, cancel := c.grpcContext(ctx, defaultGRPCExecutionTimeout)\n\tdefer cancel()\n\n\tticker := time.NewTicker(defaultRequestPollInterval)",
  "\texecCtx, cancel := c.executionContext(ctx)\n\tdefer cancel()\n\n\tticker := time.NewTicker(defaultRequestPollInterval)");

# 6. helper + ResumeStream + ExecuteToolWithKey after executeToolGRPC's end.
open(my $f, '<', $cf) or die $!; local $/; my $src = <$f>; close($f);
my $anchor = "// StreamExecuteTool consumes a live gRPC execution stream in-order.";
my $add = <<'ADD';
// executionContext derives the bounded context for an execution-lifecycle
// call: the caller's deadline when set, otherwise the configured execution
// timeout.
func (c *ToolplaneClient) executionContext(parent context.Context) (context.Context, context.CancelFunc) {
	if _, ok := parent.Deadline(); ok {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, c.executionTimeout)
}

// ExecuteToolWithKey is ExecuteTool with an idempotency key: retrying the
// same key returns the original request instead of executing the tool again.
func (c *ToolplaneClient) ExecuteToolWithKey(ctx context.Context, toolName string, params map[string]interface{}, idempotencyKey string) (*pb.Request, error) {
	if c.protocol != ProtocolGRPC {
		return nil, fmt.Errorf("tool execution only supported with gRPC protocol")
	}
	return c.executeToolGRPC(ctx, toolName, params, idempotencyKey)
}

// ResumeStream replays a request's retained chunks that follow lastSeq and
// streams live chunks until the final marker. OnChunk (optional) observes
// each chunk as it arrives; the full in-order slice is returned.
func (c *ToolplaneClient) ResumeStream(
	ctx context.Context,
	requestID string,
	lastSeq int32,
	onChunk func(*pb.ExecuteToolChunk) error,
) ([]*pb.ExecuteToolChunk, error) {
	if err := c.ensureGRPCConnected(); err != nil {
		return nil, err
	}

	streamCtx, cancel := c.executionContext(ctx)
	defer cancel()

	stream, err := c.requestsClient.ResumeStream(streamCtx, &pb.ResumeStreamRequest{
		SessionId: c.sessionID,
		RequestId: requestID,
		LastSeq:   lastSeq,
	})
	if err != nil {
		return nil, FromGRPC("resume stream", requestID, err)
	}

	chunks := make([]*pb.ExecuteToolChunk, 0, 8)
	for {
		chunk, recvErr := stream.Recv()
		if recvErr != nil {
			return chunks, FromGRPC("resume stream chunk", requestID, recvErr)
		}

		chunks = append(chunks, chunk)
		if onChunk != nil {
			if callbackErr := onChunk(chunk); callbackErr != nil {
				return chunks, callbackErr
			}
		}

		if chunk.GetIsFinal() {
			if chunk.GetError() != "" {
				return chunks, &Error{
					Op:        "resume stream",
					RequestID: requestID,
					Code:      codes.Internal,
					Message:   chunk.GetError(),
				}
			}
			return chunks, nil
		}
	}
}

ADD
die "go anchor missing" unless index($src, $anchor) >= 0;
$src =~ s/\Q$anchor\E/$add . $anchor/e;

# 7. ExecuteTool -> withKey plumbing.
my $olde = "func (c *ToolplaneClient) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (*pb.Request, error) {\n\tif c.protocol != ProtocolGRPC {\n\t\treturn nil, fmt.Errorf(\"tool execution only supported with gRPC protocol\")\n\t}\n\n\treturn c.executeToolGRPC(ctx, toolName, params)\n}";
my $newe = "func (c *ToolplaneClient) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (*pb.Request, error) {\n\treturn c.ExecuteToolWithKey(ctx, toolName, params, \"\")\n}";
die "exec wrapper missing" unless index($src, $olde) >= 0;
$src =~ s/\Q$olde\E/$newe/;

my $oldx = "func (c *ToolplaneClient) executeToolGRPC(ctx context.Context, toolName string, params map[string]interface{}) (*pb.Request, error) {";
my $newx = "func (c *ToolplaneClient) executeToolGRPC(ctx context.Context, toolName string, params map[string]interface{}, idempotencyKey string) (*pb.Request, error) {";
die "exec sig missing" unless index($src, $oldx) >= 0;
$src =~ s/\Q$oldx\E/$newx/;

my $oldr = "\trequest := &pb.ExecuteToolRequest{\n\t\tSessionId: c.sessionID,\n\t\tToolName:  toolName,\n\t\tInput:     string(paramsJSON),\n\t}\n\n\texecCtx, cancel := c.executionContext(ctx)";
my $newr = "\trequest := &pb.ExecuteToolRequest{\n\t\tSessionId:      c.sessionID,\n\t\tToolName:       toolName,\n\t\tInput:          string(paramsJSON),\n\t\tIdempotencyKey: idempotencyKey,\n\t}\n\n\texecCtx, cancel := c.executionContext(ctx)";
die "exec req missing" unless index($src, $oldr) >= 0;
$src =~ s/\Q$oldr\E/$newr/;

open(my $o, '>', $cf) or die $!; print $o $src; close($o);
$n += 4;

# 8. WithExecutionTimeout option in tls.go next to the other options.
my $tf = 'clients/go-client/client/tls.go';
edit($tf,
  "// WithGRPCTLS enables TLS for the direct gRPC client and optionally supplies a custom CA bundle and server name.",
  "// WithExecutionTimeout bounds ExecuteTool waits and execution streams.\n// The default is 30s; zero or negative keeps the default.\nfunc WithExecutionTimeout(d time.Duration) ClientOption {\n\treturn func(c *ToolplaneClient) {\n\t\tif d > 0 {\n\t\t\tc.executionTimeout = d\n\t\t}\n\t}\n}\n\n// WithGRPCTLS enables TLS for the direct gRPC client and optionally supplies a custom CA bundle and server name.");

print "applied $n\n";
