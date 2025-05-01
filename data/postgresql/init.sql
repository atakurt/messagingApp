CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    phone_number VARCHAR(20) NOT NULL,
    content VARCHAR(160) NOT NULL,
    sent BOOLEAN NOT NULL DEFAULT FALSE,
    sent_at TIMESTAMP,
    message_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_messages_sent ON messages (sent);

INSERT INTO messages (phone_number, content, sent)
VALUES
  ('+905321234567', 'Hey whats up ?', false),
  ('+905321234568', 'Reminder: Call tomorrow.', false),
  ('+905321234567', 'call me as soon as possible !', false),
  ('+905321234567', 'call me as soon as possible !', false),
  ('+905321234567', 'call me as soon as possible !', false),
  ('+905321234567', 'call me as soon as possible !', false),
  ('+905321234567', 'call me as soon as possible !', false),
  ('+905321234567', 'call me as soon as possible !', false),
  ('+905321234567', 'call me as soon as possible !', false),
  ('+905321234567', 'call me as soon as possible !', false),
  ('+905321234567', 'call me as soon as possible !', false),
  ('+905321234567', 'call me as soon as possible !', false),
  ('+905321234567', 'call me as soon as possible !', false);
