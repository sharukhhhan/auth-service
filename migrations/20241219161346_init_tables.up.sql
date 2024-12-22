CREATE TABLE IF NOT EXISTS users (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       email VARCHAR(255) NOT NULL UNIQUE,
                       created_at TIMESTAMP DEFAULT NOW(),
                       updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                refresh_hash VARCHAR(255) NOT NULL,
                                issued_at TIMESTAMP DEFAULT NOW(),
                                expires_at TIMESTAMP NOT NULL,
                                client_ip VARCHAR(255) NOT NULL,
                                used BOOLEAN DEFAULT FALSE,
                                UNIQUE(user_id, refresh_hash)
);