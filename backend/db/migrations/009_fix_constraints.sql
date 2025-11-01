-- Migration 009: Fix constraint names for GORM compatibility
-- GORM expects specific constraint naming conventions (uni_* prefix)

-- Fix calendars table constraints
ALTER TABLE calendars DROP CONSTRAINT IF EXISTS calendars_recall_calendar_id_key;
ALTER TABLE calendars DROP CONSTRAINT IF EXISTS uni_calendars_recall_calendar_id;
ALTER TABLE calendars ADD CONSTRAINT uni_calendars_recall_calendar_id UNIQUE (recall_calendar_id);

-- Fix calendar_events table constraints
ALTER TABLE calendar_events DROP CONSTRAINT IF EXISTS calendar_events_recall_event_id_key;
ALTER TABLE calendar_events DROP CONSTRAINT IF EXISTS uni_calendar_events_recall_event_id;
ALTER TABLE calendar_events ADD CONSTRAINT uni_calendar_events_recall_event_id UNIQUE (recall_event_id);

-- Fix calendar_o_auth_states table constraints (if any)
-- Check for state column unique constraint
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'calendar_o_auth_states_state_key') THEN
        ALTER TABLE calendar_o_auth_states DROP CONSTRAINT calendar_o_auth_states_state_key;
    END IF;
END $$;
ALTER TABLE calendar_o_auth_states DROP CONSTRAINT IF EXISTS uni_calendar_o_auth_states_state;
ALTER TABLE calendar_o_auth_states ADD CONSTRAINT uni_calendar_o_auth_states_state UNIQUE (state);

-- Ensure all tables have proper structure after users table removal
-- This is idempotent and safe to run multiple times

