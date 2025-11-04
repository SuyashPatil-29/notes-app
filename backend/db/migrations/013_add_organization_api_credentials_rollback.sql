-- Rollback migration: Remove organization API credentials table
-- This will drop the organization_api_credentials table and all its data

-- Drop indexes first
DROP INDEX IF EXISTS idx_org_api_credentials_created_by;
DROP INDEX IF EXISTS idx_org_api_credentials_org_provider;
DROP INDEX IF EXISTS idx_org_api_credentials_provider;
DROP INDEX IF EXISTS idx_org_api_credentials_organization_id;

-- Drop the table
DROP TABLE IF EXISTS organization_api_credentials;