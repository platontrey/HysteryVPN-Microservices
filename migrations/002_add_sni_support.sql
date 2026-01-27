-- Migration: Add SNI support to vps_nodes table
-- Description: Add fields for SNI domain management and certificate storage
-- Version: 002

-- Add SNI-related columns to vps_nodes table
ALTER TABLE vps_nodes 
ADD COLUMN primary_domain VARCHAR(255),
ADD COLUMN sni_enabled BOOLEAN DEFAULT FALSE,
ADD COLUMN certificate_path VARCHAR(500),
ADD COLUMN key_path VARCHAR(500),
ADD COLUMN sni_auto_renew BOOLEAN DEFAULT TRUE,
ADD COLUMN sni_email VARCHAR(255);

-- Update existing metadata to include empty SNI domains array for backward compatibility
UPDATE vps_nodes 
SET metadata = COALESCE(metadata, '{}'::jsonb) || '{"sni_domains": []}'::jsonb
WHERE metadata IS NULL OR NOT metadata ? 'sni_domains';

-- Create index for better performance on SNI-related queries
CREATE INDEX idx_vps_nodes_sni_enabled ON vps_nodes(sni_enabled) WHERE sni_enabled = TRUE;
CREATE INDEX idx_vps_nodes_primary_domain ON vps_nodes(primary_domain) WHERE primary_domain IS NOT NULL;

-- Add comments to describe new columns
COMMENT ON COLUMN vps_nodes.primary_domain IS 'Primary domain for SNI configuration';
COMMENT ON COLUMN vps_nodes.sni_enabled IS 'Whether SNI is enabled for this node';
COMMENT ON COLUMN vps_nodes.certificate_path IS 'Path to SSL certificate file';
COMMENT ON COLUMN vps_nodes.key_path IS 'Path to SSL private key file';
COMMENT ON COLUMN vps_nodes.sni_auto_renew IS 'Whether to automatically renew certificates';
COMMENT ON COLUMN vps_nodes.sni_email IS 'Email address for certificate notifications (Let''s Encrypt)';

-- Create trigger to ensure data consistency
CREATE OR REPLACE FUNCTION validate_sni_configuration()
RETURNS TRIGGER AS $$
BEGIN
    -- If SNI is enabled, ensure primary domain is set
    IF NEW.sni_enabled = TRUE AND NEW.primary_domain IS NULL THEN
        RAISE EXCEPTION 'Primary domain must be set when SNI is enabled';
    END IF;
    
    -- Ensure certificate and key paths are set when SNI is enabled
    IF NEW.sni_enabled = TRUE THEN
        IF NEW.certificate_path IS NULL THEN
            NEW.certificate_path = '/etc/hysteria/sni/' || NEW.primary_domain || '.crt';
        END IF;
        
        IF NEW.key_path IS NULL THEN
            NEW.key_path = '/etc/hysteria/sni/' || NEW.primary_domain || '.key';
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_validate_sni_configuration
    BEFORE INSERT OR UPDATE ON vps_nodes
    FOR EACH ROW
    EXECUTE FUNCTION validate_sni_configuration();

-- Create function to extract SNI domains from metadata
CREATE OR REPLACE FUNCTION get_sni_domains(node_metadata jsonb)
RETURNS text[] AS $$
BEGIN
    RETURN COALESCE(
        ARRAY(
            SELECT value::text
            FROM jsonb_array_elements_text(COALESCE(node_metadata->'sni_domains', '[]'::jsonb)) AS value
        ),
        ARRAY[]::text[]
    );
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Create view for easy SNI management
CREATE OR REPLACE VIEW sni_nodes_view AS
SELECT 
    id,
    name,
    hostname,
    ip_address,
    location,
    country,
    status,
    primary_domain,
    sni_enabled,
    certificate_path,
    key_path,
    sni_auto_renew,
    sni_email,
    get_sni_domains(metadata) as sni_domains,
    last_heartbeat,
    created_at
FROM vps_nodes
WHERE sni_enabled = TRUE;

-- Grant necessary permissions (adjust role names as needed)
-- GRANT SELECT, UPDATE ON vps_nodes TO hysteria_api;
-- GRANT SELECT ON sni_nodes_view TO hysteria_api;
-- GRANT EXECUTE ON FUNCTION get_sni_domains TO hysteria_api;

-- Log migration completion
DO $$
BEGIN
    RAISE NOTICE 'Migration 002: SNI support completed successfully';
    RAISE NOTICE 'Added SNI columns to vps_nodes table';
    RAISE NOTICE 'Created indexes and validation triggers';
    RAISE NOTICE 'Created sni_nodes_view for SNI management';
END $$;