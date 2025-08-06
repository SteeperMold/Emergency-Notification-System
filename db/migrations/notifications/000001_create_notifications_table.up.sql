CREATE TYPE notification_status AS ENUM (
    'pending',
    'in_flight',
    'sent',
    'failed'
    );

CREATE TABLE notifications
(
    id              uuid PRIMARY KEY,
    user_id         INT                 NOT NULL,
    text            TEXT                NOT NULL,
    recipient_phone TEXT                NOT NULL,
    status          notification_status NOT NULL DEFAULT 'pending',
    attempts        INT                 NOT NULL DEFAULT 0,
    next_run_at     TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_pending
    ON notifications (next_run_at, status)
    WHERE status = 'pending';
