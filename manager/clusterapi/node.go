package clusterapi

import (
	"github.com/docker/swarm-v2/api"
	"github.com/docker/swarm-v2/manager/state"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func validateNodeSpec(spec *api.NodeSpec) error {
	if spec == nil {
		return grpc.Errorf(codes.InvalidArgument, errInvalidArgument.Error())
	}
	return nil
}

// GetNode returns a Node given a NodeID.
// - Returns `InvalidArgument` if NodeID is not provided.
// - Returns `NotFound` if the Node is not found.
func (s *Server) GetNode(ctx context.Context, request *api.GetNodeRequest) (*api.GetNodeResponse, error) {
	if request.NodeID == "" {
		return nil, grpc.Errorf(codes.InvalidArgument, errInvalidArgument.Error())
	}

	var node *api.Node
	err := s.store.View(func(tx state.ReadTx) error {
		node = tx.Nodes().Get(request.NodeID)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, grpc.Errorf(codes.NotFound, "node %s not found", request.NodeID)
	}
	return &api.GetNodeResponse{
		Node: node,
	}, nil
}

// ListNodes returns a list of all nodes.
func (s *Server) ListNodes(ctx context.Context, request *api.ListNodesRequest) (*api.ListNodesResponse, error) {
	var nodes []*api.Node
	err := s.store.View(func(tx state.ReadTx) error {
		var err error
		if request.Options == nil || request.Options.Query == "" {
			nodes, err = tx.Nodes().Find(state.All)
		} else {
			nodes, err = tx.Nodes().Find(state.ByQuery(request.Options.Query))
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return &api.ListNodesResponse{
		Nodes: nodes,
	}, nil
}

// UpdateNode updates a Node referenced by NodeID with the given NodeSpec.
// - Returns `NotFound` if the Node is not found.
// - Returns `InvalidArgument` if the NodeSpec is malformed.
// - Returns an error if the update fails.
func (s *Server) UpdateNode(ctx context.Context, request *api.UpdateNodeRequest) (*api.UpdateNodeResponse, error) {
	if request.NodeID == "" || request.NodeVersion == nil {
		return nil, grpc.Errorf(codes.InvalidArgument, errInvalidArgument.Error())
	}
	if err := validateNodeSpec(request.Spec); err != nil {
		return nil, err
	}

	var node *api.Node
	err := s.store.Update(func(tx state.Tx) error {
		node = tx.Nodes().Get(request.NodeID)
		if node == nil {
			return nil
		}
		node.Version = *request.NodeVersion
		node.Spec = *request.Spec.Copy()
		return tx.Nodes().Update(node)
	})
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, grpc.Errorf(codes.NotFound, "node %s not found", request.NodeID)
	}
	return &api.UpdateNodeResponse{
		Node: node,
	}, nil
}
