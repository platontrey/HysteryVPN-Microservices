package handlers

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"hysteria2_microservices/orchestrator-service/internal/models"
	pb "hysteria2_microservices/proto"
)

// NodeConfigHandler handles SNI configuration for nodes
type NodeConfigHandler struct {
	nodeHandler *NodeHandler
}

// NewNodeConfigHandler creates a new NodeConfigHandler
func NewNodeConfigHandler(nodeHandler *NodeHandler) *NodeConfigHandler {
	return &NodeConfigHandler{
		nodeHandler: nodeHandler,
	}
}

// GetSNIConfig retrieves SNI configuration for a node
func (h *NodeConfigHandler) GetSNIConfig(ctx context.Context, nodeID string) (*models.VPSNode, error) {
	// Get node from database
	var node models.VPSNode
	if err := h.nodeHandler.db.First(&node, "id = ?", nodeID).Error; err != nil {
		return nil, fmt.Errorf("node not found: %w", err)
	}

	return &node, nil
}

// UpdateSNIConfig updates SNI configuration for a node
func (h *NodeConfigHandler) UpdateSNIConfig(ctx context.Context, req *pb.UpdateSNIConfigRequest) (*pb.UpdateSNIConfigResponse, error) {
	// Get node from database
	var node models.VPSNode
	if err := h.nodeHandler.db.First(&node, "id = ?", req.NodeId).Error; err != nil {
		return nil, fmt.Errorf("node not found: %w", err)
	}

	// Update SNI fields
	node.SNIEnabled = req.SniEnabled
	node.PrimaryDomain = req.PrimaryDomain
	node.SNIAutoRenew = req.AutoRenew
	node.SNIEmail = req.Email

	// Update SNI domains
	node.SetSNIDomains(req.Domains)

	// Save to database
	if err := h.nodeHandler.db.Save(&node).Error; err != nil {
		return nil, fmt.Errorf("failed to update node: %w", err)
	}

	// Connect to node via gRPC and update configuration
	conn, err := h.nodeHandler.getNodeConnection(req.NodeId)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to node: %w", err)
	}
	defer conn.Close()

	client := pb.NewNodeManagerClient(conn)

	// Generate and send new configuration with SNI
	configUpdate := &pb.SNIConfigUpdateRequest{
		Enabled:    req.SniEnabled,
		Domains:    req.Domains,
		DefaultSni: req.PrimaryDomain,
		AutoMode:   true, // Always use auto mode for simplicity
	}

	_, err = client.UpdateSNIConfig(ctx, configUpdate)
	if err != nil {
		return nil, fmt.Errorf("failed to update SNI config on node: %w", err)
	}

	return &pb.UpdateSNIConfigResponse{
		Success: true,
		Message: "SNI configuration updated successfully",
	}, nil
}

// AddSNIDomain adds a new domain to node's SNI configuration
func (h *NodeConfigHandler) AddSNIDomain(ctx context.Context, nodeID, domain string) error {
	// Get node from database
	var node models.VPSNode
	if err := h.nodeHandler.db.First(&node, "id = ?", nodeID).Error; err != nil {
		return fmt.Errorf("node not found: %w", err)
	}

	// Check if domain already exists
	if node.HasSNIDomain(domain) {
		return fmt.Errorf("domain %s already exists", domain)
	}

	// Add domain
	node.AddSNIDomain(domain)

	// If this is the first domain and SNI is not enabled, enable it
	if len(node.GetSNIDomains()) == 1 && !node.SNIEnabled {
		node.SNIEnabled = true
		node.PrimaryDomain = domain
	}

	// Save to database
	if err := h.nodeHandler.db.Save(&node).Error; err != nil {
		return fmt.Errorf("failed to update node: %w", err)
	}

	// Update configuration on node
	conn, err := h.nodeHandler.getNodeConnection(nodeID)
	if err != nil {
		return fmt.Errorf("failed to connect to node: %w", err)
	}
	defer conn.Close()

	client := pb.NewNodeManagerClient(conn)

	_, err = client.AddSNIDomain(ctx, &pb.AddSNIDomainRequest{
		Domain: domain,
	})
	if err != nil {
		return fmt.Errorf("failed to add domain on node: %w", err)
	}

	return nil
}

// RemoveSNIDomain removes a domain from node's SNI configuration
func (h *NodeConfigHandler) RemoveSNIDomain(ctx context.Context, nodeID, domain string) error {
	// Get node from database
	var node models.VPSNode
	if err := h.nodeHandler.db.First(&node, "id = ?", nodeID).Error; err != nil {
		return fmt.Errorf("node not found: %w", err)
	}

	// Remove domain
	if !node.HasSNIDomain(domain) {
		return fmt.Errorf("domain %s not found", domain)
	}

	node.RemoveSNIDomain(domain)

	// Update primary domain if necessary
	if node.PrimaryDomain == domain {
		domains := node.GetSNIDomains()
		if len(domains) > 0 {
			node.PrimaryDomain = domains[0]
		} else {
			node.SNIEnabled = false
			node.PrimaryDomain = ""
		}
	}

	// Save to database
	if err := h.nodeHandler.db.Save(&node).Error; err != nil {
		return fmt.Errorf("failed to update node: %w", err)
	}

	// Update configuration on node
	conn, err := h.nodeHandler.getNodeConnection(nodeID)
	if err != nil {
		return fmt.Errorf("failed to connect to node: %w", err)
	}
	defer conn.Close()

	client := pb.NewNodeManagerClient(conn)

	_, err = client.RemoveSNIDomain(ctx, &pb.RemoveSNIDomainRequest{
		Domain: domain,
	})
	if err != nil {
		return fmt.Errorf("failed to remove domain on node: %w", err)
	}

	return nil
}

// GetSNIStatus retrieves SNI status for a node
func (h *NodeConfigHandler) GetSNIStatus(ctx context.Context, nodeID string) (*pb.SNIStatusResponse, error) {
	// Get node from database
	var node models.VPSNode
	if err := h.nodeHandler.db.First(&node, "id = ?", nodeID).Error; err != nil {
		return nil, fmt.Errorf("node not found: %w", err)
	}

	// Connect to node to get certificate status
	conn, err := h.nodeHandler.getNodeConnection(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to node: %w", err)
	}
	defer conn.Close()

	client := pb.NewNodeManagerClient(conn)

	statusResp, err := client.GetSNIStatus(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get SNI status from node: %w", err)
	}

	return &pb.SNIStatusResponse{
		Enabled:       node.SNIEnabled,
		Domains:       node.GetSNIDomains(),
		PrimaryDomain: node.PrimaryDomain,
		AutoRenew:     node.SNIAutoRenew,
		Email:         node.SNIEmail,
		Certificates:  statusResp.Certificates,
		ExpiringSoon:  statusResp.ExpiringSoon,
	}, nil
}
