-- Create calendars table
CREATE TABLE IF NOT EXISTS calendars (
    id VARCHAR(255) PRIMARY KEY,
    user_id INTEGER NOT NULL,
    recall_calendar_id VARCHAR(255) UNIQUE NOT NULL,
    platform VARCHAR(50) NOT NULL,
    platform_email VARCHAR(255) NOT NULL,
    oauth_client_id TEXT NOT NULL,
    oauth_client_secret TEXT NOT NULL,
    oauth_refresh_token TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    last_synced_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create calendar_events table
CREATE TABLE IF NOT EXISTS calendar_events (
    id VARCHAR(255) PRIMARY KEY,
    calendar_id VARCHAR(255) NOT NULL,
    recall_event_id VARCHAR(255) UNIQUE NOT NULL,
    i_cal_uid VARCHAR(255),
    platform_id VARCHAR(255),
    meeting_platform VARCHAR(100),
    meeting_url TEXT,
    title TEXT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE,
    bot_scheduled BOOLEAN DEFAULT FALSE,
    bot_id VARCHAR(255),
    meeting_recording_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (calendar_id) REFERENCES calendars(id) ON DELETE CASCADE,
    FOREIGN KEY (meeting_recording_id) REFERENCES meeting_recordings(id) ON DELETE SET NULL
);

-- Create calendar_o_auth_states table for OAuth flow security (note: GORM naming convention)
CREATE TABLE IF NOT EXISTS calendar_o_auth_states (
    id VARCHAR(255) PRIMARY KEY,
    user_id INTEGER NOT NULL,
    state VARCHAR(255) UNIQUE NOT NULL,
    platform VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_calendars_user_id ON calendars(user_id);
CREATE INDEX IF NOT EXISTS idx_calendars_recall_calendar_id ON calendars(recall_calendar_id);
CREATE INDEX IF NOT EXISTS idx_calendar_events_calendar_id ON calendar_events(calendar_id);
CREATE INDEX IF NOT EXISTS idx_calendar_events_recall_event_id ON calendar_events(recall_event_id);
CREATE INDEX IF NOT EXISTS idx_calendar_events_start_time ON calendar_events(start_time);
CREATE INDEX IF NOT EXISTS idx_calendar_o_auth_states_user_id ON calendar_o_auth_states(user_id);
CREATE INDEX IF NOT EXISTS idx_calendar_o_auth_states_state ON calendar_o_auth_states(state);
CREATE INDEX IF NOT EXISTS idx_calendar_o_auth_states_expires_at ON calendar_o_auth_states(expires_at);

