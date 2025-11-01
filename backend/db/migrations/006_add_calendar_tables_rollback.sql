-- Drop indexes
DROP INDEX IF EXISTS idx_calendar_o_auth_states_expires_at;
DROP INDEX IF EXISTS idx_calendar_o_auth_states_state;
DROP INDEX IF EXISTS idx_calendar_o_auth_states_user_id;
DROP INDEX IF EXISTS idx_calendar_events_start_time;
DROP INDEX IF EXISTS idx_calendar_events_recall_event_id;
DROP INDEX IF EXISTS idx_calendar_events_calendar_id;
DROP INDEX IF EXISTS idx_calendars_recall_calendar_id;
DROP INDEX IF EXISTS idx_calendars_user_id;

-- Drop tables
DROP TABLE IF EXISTS calendar_o_auth_states;
DROP TABLE IF EXISTS calendar_events;
DROP TABLE IF EXISTS calendars;

