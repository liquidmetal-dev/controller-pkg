package client

import (
	"context"

	flintlockv1 "github.com/liquidmetal-dev/flintlock/api/services/microvm/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type flintlockClient struct {
	c    flintlockv1.MicroVMClient
	conn *grpc.ClientConn
}

func (fc *flintlockClient) Close() {
	if fc.conn != nil {
		fc.conn.Close()
	}
}

func (fc *flintlockClient) CreateMicroVM(ctx context.Context, in *flintlockv1.CreateMicroVMRequest, opts ...grpc.CallOption) (*flintlockv1.CreateMicroVMResponse, error) { //nolint:lll // it would make it less readable
	return fc.c.CreateMicroVM(ctx, in, opts...)
}

func (fc *flintlockClient) DeleteMicroVM(ctx context.Context, in *flintlockv1.DeleteMicroVMRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) { //nolint:lll // it would make it less readable
	return fc.c.DeleteMicroVM(ctx, in, opts...)
}

func (fc *flintlockClient) GetMicroVM(ctx context.Context, in *flintlockv1.GetMicroVMRequest, opts ...grpc.CallOption) (*flintlockv1.GetMicroVMResponse, error) { //nolint:lll // it would make it less readable
	return fc.c.GetMicroVM(ctx, in, opts...)
}

func (fc *flintlockClient) ListMicroVMs(ctx context.Context, in *flintlockv1.ListMicroVMsRequest, opts ...grpc.CallOption) (*flintlockv1.ListMicroVMsResponse, error) { //nolint:lll // it would make it less readable
	return fc.c.ListMicroVMs(ctx, in, opts...)
}

func (fc *flintlockClient) ListMicroVMsStream(ctx context.Context, in *flintlockv1.ListMicroVMsRequest, opts ...grpc.CallOption) (flintlockv1.MicroVM_ListMicroVMsStreamClient, error) { //nolint:lll // it would make it less readable
	return fc.c.ListMicroVMsStream(ctx, in, opts...)
}
