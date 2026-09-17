package cli

import (
	"context"
	"time"

	"google.golang.org/grpc/metadata"

	proto "toolplane/proto"
)

// DefaultReadyPollInterval is the cadence WaitReady polls at.
const DefaultReadyPollInterval = 250 * time.Millisecond

// WaitReady polls the api.v1 HealthCheck until the server answers or the
// context is done, returning the last probe error. Used by
// `toolplane wait --for ready` and by scripts that need to block on
// server availability before acting.
func WaitReady(ctx context.Context, conn Connection, interval time.Duration) error {
	if interval <= 0 {
		interval = DefaultReadyPollInterval
	}
	gconn, err := conn.Dial()
	if err != nil {
		return err
	}
	defer gconn.Close()
	tool := proto.NewToolServiceClient(gconn)

	var lastErr error
	for {
		probeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		if conn.APIKey != "" {
			probeCtx = metadata.AppendToOutgoingContext(probeCtx, "api_key", conn.APIKey)
		}
		_, pingErr := tool.HealthCheck(probeCtx, &proto.HealthCheckRequest{})
		cancel()
		if pingErr == nil {
			return nil
		}
		lastErr = pingErr
		if ctx.Err() != nil {
			return lastErr
		}
		time.Sleep(interval)
	}
}
