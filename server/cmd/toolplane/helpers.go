package main

import (
	"context"

	"google.golang.org/grpc"

	"toolplane/internal/cli"
	proto "toolplane/proto"
)

// dialTools opens the control-plane connection for the ToolService.
func dialTools(conn cli.Connection) (proto.ToolServiceClient, context.Context, context.CancelFunc, *grpc.ClientConn, error) {
	gconn, err := conn.Dial()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	ctx, cancel := conn.Call(context.Background())
	return proto.NewToolServiceClient(gconn), ctx, cancel, gconn, nil
}

// dialMachines opens the control-plane connection for MachinesService.
func dialMachines(conn cli.Connection) (proto.MachinesServiceClient, context.Context, context.CancelFunc, *grpc.ClientConn, error) {
	gconn, err := conn.Dial()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	ctx, cancel := conn.Call(context.Background())
	return proto.NewMachinesServiceClient(gconn), ctx, cancel, gconn, nil
}
