CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    phone_number VARCHAR(20) NOT NULL,
    content VARCHAR(160) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    message_id VARCHAR(255),
    last_error TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    sent_at TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_messages_status ON messages (status);

INSERT INTO messages (phone_number, content, status)
VALUES
  ('+905321234567', 'Hey whats up ?', 'pending'),
  ('+905321234568', 'Reminder: Call tomorrow.', 'pending');


INSERT INTO messages (phone_number, content, status)
SELECT
    '+905' || LPAD((FLOOR(RANDOM() * 100000000)::int)::text, 8, '0'),
    'Random message ' || i,
    'pending'
FROM generate_series(1, 62) AS s(i);


CREATE TABLE message_retries (
                                 id SERIAL PRIMARY KEY,
                                 original_message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
                                 phone_number TEXT NOT NULL,
                                 content TEXT NOT NULL,
                                 retry_count INT NOT NULL DEFAULT 0,
                                 last_error TEXT,
                                 created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_message_retries_original_message_id ON message_retries(original_message_id);


CREATE TABLE message_dead_letters (
                                      id SERIAL PRIMARY KEY,
                                      original_message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
                                      phone_number TEXT NOT NULL,
                                      content TEXT NOT NULL,
                                      last_error TEXT,
                                      failed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_message_dead_letters_original_message_id ON message_dead_letters(original_message_id);
CREATE INDEX idx_message_dead_letters_failed_at ON message_dead_letters(failed_at);


