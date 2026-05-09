CREATE TABLE revoked_tokens (
    jti UUID PRIMARY KEY,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_revoked_tokens_expires_at ON revoked_tokens(expires_at);
