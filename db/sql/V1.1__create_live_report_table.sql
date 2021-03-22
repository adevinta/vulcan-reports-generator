CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE live_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_subject TEXT NOT NULL DEFAULT '',
    email_body TEXT NOT NULL DEFAULT '',
    team_id TEXT NOT NULL DEFAULT '',
    date_to TEXT NOT NULL DEFAULT '',
    date_from TEXT NOT NULL DEFAULT '',
    delivered_to TEXT NOT NULL DEFAULT '',
    update_status_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(team_id, date_from, date_to)
);
