-- Migration: Add organization API credentials table
-- This table stores API keys configured at the organization level
-- Organization admins can set these keys to be used by all organization members

CREATE TABLE organization_api_credentials (
    id SERIAL PRIMARY KEY,
    organization_id VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    key_cipher BYTEA NOT NULL,
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure unique combination of organization and provider
    UNIQUE(organization_id, provider)
);

-- Create indexes for better query performance
CREATE INDEX idx_org_api_credentials_organization_id ON organization_api_credentials(organization_id);
CREATE INDEX idx_org_api_credentials_provider ON organization_api_credentials(provider);
CREATE INDEX idx_org_api_credentials_org_provider ON organization_api_credentials(organization_id, provider);
CREATE INDEX idx_org_api_credentials_created_by ON organization_api_credentials(created_by);

-- Add comments for documentation
COMMENT ON TABLE organization_api_credentials IS 'Stores API credentials configured at the organization level for AI services';
COMMENT ON COLUMN organization_api_credentials.organization_id IS 'Clerk organization ID';
COMMENT ON COLUMN organization_api_credentials.provider IS 'AI provider name (openai, anthropic, google, etc.)';
COMMENT ON COLUMN organization_api_credentials.key_cipher IS 'AES-GCM encrypted API key';
COMMENT ON COLUMN organization_api_credentials.created_by IS 'Clerk user ID of the admin who created this credential';