package server

import (
	"context"
	"testing"

	entitygraphpb "github.com/aosanya/CodeValdSharedLib/gen/go/entitygraph/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTraverseDepthInterceptor_BelowLimit(t *testing.T) {
	interceptor := TraverseDepthInterceptor(10)
	info := &grpc.UnaryServerInfo{FullMethod: entitygraphpb.EntityService_TraverseGraph_FullMethodName}
	req := &entitygraphpb.TraverseGraphRequest{Depth: 5}

	called := false
	handler := func(_ context.Context, _ any) (any, error) { called = true; return nil, nil }

	_, err := interceptor(context.Background(), req, info, handler)
	if err != nil {
		t.Fatalf("expected no error for depth=5, got: %v", err)
	}
	if !called {
		t.Error("expected handler to be called")
	}
}

func TestTraverseDepthInterceptor_AtLimit(t *testing.T) {
	interceptor := TraverseDepthInterceptor(10)
	info := &grpc.UnaryServerInfo{FullMethod: entitygraphpb.EntityService_TraverseGraph_FullMethodName}
	req := &entitygraphpb.TraverseGraphRequest{Depth: 10}

	called := false
	handler := func(_ context.Context, _ any) (any, error) { called = true; return nil, nil }

	_, err := interceptor(context.Background(), req, info, handler)
	if err != nil {
		t.Fatalf("expected no error for depth=10 (at limit), got: %v", err)
	}
	if !called {
		t.Error("expected handler to be called at limit")
	}
}

func TestTraverseDepthInterceptor_ExceedsLimit(t *testing.T) {
	interceptor := TraverseDepthInterceptor(10)
	info := &grpc.UnaryServerInfo{FullMethod: entitygraphpb.EntityService_TraverseGraph_FullMethodName}
	req := &entitygraphpb.TraverseGraphRequest{Depth: 11}

	called := false
	handler := func(_ context.Context, _ any) (any, error) { called = true; return nil, nil }

	_, err := interceptor(context.Background(), req, info, handler)
	if err == nil {
		t.Fatal("expected error for depth=11, got nil")
	}
	if called {
		t.Error("handler should not be called when depth exceeds limit")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %T", err)
	}
	if st.Code() != codes.InvalidArgument {
		t.Errorf("expected codes.InvalidArgument, got %v", st.Code())
	}
}

func TestTraverseDepthInterceptor_ZeroDepth(t *testing.T) {
	interceptor := TraverseDepthInterceptor(10)
	info := &grpc.UnaryServerInfo{FullMethod: entitygraphpb.EntityService_TraverseGraph_FullMethodName}
	req := &entitygraphpb.TraverseGraphRequest{Depth: 0}

	called := false
	handler := func(_ context.Context, _ any) (any, error) { called = true; return nil, nil }

	_, err := interceptor(context.Background(), req, info, handler)
	if err != nil {
		t.Fatalf("expected no error for depth=0 (default), got: %v", err)
	}
	if !called {
		t.Error("expected handler to be called for depth=0")
	}
}

func TestTraverseDepthInterceptor_OtherMethod_Unaffected(t *testing.T) {
	interceptor := TraverseDepthInterceptor(10)
	info := &grpc.UnaryServerInfo{FullMethod: "/entitygraph.v1.EntityService/GetEntity"}
	req := &entitygraphpb.GetEntityRequest{AgencyId: "a", EntityId: "b"}

	called := false
	handler := func(_ context.Context, _ any) (any, error) { called = true; return nil, nil }

	_, err := interceptor(context.Background(), req, info, handler)
	if err != nil {
		t.Fatalf("expected no error for non-TraverseGraph method, got: %v", err)
	}
	if !called {
		t.Error("expected handler to be called for non-TraverseGraph method")
	}
}
