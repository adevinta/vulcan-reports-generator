CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE scan_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scan_id TEXT NOT NULL,
    report TEXT NOT NULL DEFAULT '',
    report_json TEXT NOT NULL DEFAULT '',
    email_subject TEXT NOT NULL DEFAULT '',
    email_body TEXT NOT NULL DEFAULT '',
    delivered_to TEXT NOT NULL DEFAULT '',
    update_status_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    program_name TEXT NOT NULL,
    risk INTEGER NOT NULL DEFAULT 0,
    UNIQUE(scan_id)
);
