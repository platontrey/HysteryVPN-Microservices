-- VPS nodes registry
CREATE TABLE IF NOT EXISTS vps_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    ip_address INET NOT NULL,
    location VARCHAR(100),
    country VARCHAR(2),
    grpc_port INTEGER DEFAULT 50051,
    status VARCHAR(20) DEFAULT 'offline' CHECK (status IN ('online', 'offline', 'maintenance', 'error')),
    version VARCHAR(50),
    capabilities JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    metadata JSONB
);

-- Node assignments
CREATE TABLE IF NOT EXISTS node_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    node_id UUID NOT NULL REFERENCES vps_nodes(id) ON DELETE CASCADE,
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true,
    UNIQUE(user_id, node_id)
);

-- Node metrics
CREATE TABLE IF NOT EXISTS node_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id UUID NOT NULL REFERENCES vps_nodes(id) ON DELETE CASCADE,
    cpu_usage DECIMAL(5,2),
    memory_usage DECIMAL(5,2),
    bandwidth_up BIGINT,
    bandwidth_down BIGINT,
    active_connections INTEGER,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Deployment history
CREATE TABLE IF NOT EXISTS deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id UUID NOT NULL REFERENCES vps_nodes(id) ON DELETE CASCADE,
    config_version VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'deploying', 'success', 'failed')),
    deployed_at TIMESTAMP WITH TIME ZONE,
    rollback_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_vps_nodes_ip_address ON vps_nodes(ip_address);
CREATE INDEX IF NOT EXISTS idx_vps_nodes_status ON vps_nodes(status);
CREATE INDEX IF NOT EXISTS idx_vps_nodes_last_heartbeat ON vps_nodes(last_heartbeat);
CREATE INDEX IF NOT EXISTS idx_node_assignments_user_id ON node_assignments(user_id);
CREATE INDEX IF NOT EXISTS idx_node_assignments_node_id ON node_assignments(node_id);
CREATE INDEX IF NOT EXISTS idx_node_metrics_node_id ON node_metrics(node_id);
CREATE INDEX IF NOT EXISTS idx_node_metrics_recorded_at ON node_metrics(recorded_at);
CREATE INDEX IF NOT EXISTS idx_deployments_node_id ON deployments(node_id);