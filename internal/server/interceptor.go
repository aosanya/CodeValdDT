package server

import (
	"context"

	entitygraphpb "github.com/aosanya/CodeValdSharedLib/gen/go/entitygraph/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MaxTraverseDepth is the ceiling enforced on TraverseGraph requests.
// Unbounded traversals can exhaust ArangoDB resources on large graphs.
const MaxTraverseDepth int32 = 10

// TraverseDepthInterceptor returns a gRPC UnaryServerInterceptor that rejects
// TraverseGraph requests whose Depth field exceeds maxDepth with
// codes.InvalidArgument. All other RPC methods pass through unmodified.
func TraverseDepthInterceptor(maxDepth int32) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if info.FullMethod == entitygraphpb.EntityService_TraverseGraph_FullMethodName {
			tReq, ok := req.(*entitygraphpb.TraverseGraphRequest)
			if ok && tReq.Depth > maxDepth {
				return nil, status.Errorf(
					codes.InvalidArgument,
					"traverse depth %d exceeds maximum of %d", tReq.Depth, maxDepth,
				)
			}
		}
		return handler(ctx, req)
	}
}
