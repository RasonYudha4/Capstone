CREATE TYPE user_role AS ENUM ('master-admin', 'admin', 'staff');
CREATE TYPE action_type AS ENUM ('insert', 'open', 'edit','update','delete','error','login','login_fail','lockout');
CREATE TYPE audit_source AS ENUM ('client', 'system');

CREATE TABLE IF NOT EXISTS groups (
    group_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_name VARCHAR(255),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID,
    email VARCHAR(255),
    verified BOOLEAN,
    role user_role,
    verification_code VARCHAR(6),
    password_hash     VARCHAR(255),
    failed_attempts   INT DEFAULT 0,
    locked_until      TIMESTAMP,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(group_id)
);

CREATE TABLE IF NOT EXISTS services (
    service_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID,
    service_code VARCHAR(20),
    description VARCHAR(2052),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(group_id)
);

CREATE TABLE IF NOT EXISTS standard (
    standard_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID,
    standard_code VARCHAR(20),
    description VARCHAR(2052),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (service_id) REFERENCES services(service_id)
);

CREATE TABLE IF NOT EXISTS assessment (
    assessment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    standard_id UUID,
    assessment_code VARCHAR(20),
    description VARCHAR(2052),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (standard_id) REFERENCES standard(standard_id)
);

CREATE TABLE IF NOT EXISTS document_types (
    document_type_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255),
    description VARCHAR(255),
    is_public BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS documents (
    document_id UUID PRIMARY KEY DEFAULT gen_random_uuid() ,
    assessment_id UUID,
    filename VARCHAR(255) UNIQUE,
    filepath VARCHAR(5024),
    filehash VARCHAR(255),
    object_id UUID,
    document_type_id UUID,
    status VARCHAR(20),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    created_by UUID,
    group_id UUID,
    service_id UUID,
    standard_id UUID,
    is_deleted BOOLEAN,
    FOREIGN KEY (assessment_id) REFERENCES assessment(assessment_id),
    FOREIGN KEY (document_type_id) REFERENCES document_types(document_type_id),
    FOREIGN KEY (created_by) REFERENCES users(user_id),
    FOREIGN KEY (group_id) REFERENCES groups(group_id),
    FOREIGN KEY (service_id) REFERENCES services(service_id),
    FOREIGN KEY (standard_id) REFERENCES standard(standard_id)
);

CREATE TABLE IF NOT EXISTS audit (
    audit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action action_type,
    description VARCHAR(255),
    user_id UUID,
    document_id UUID,
    source audit_source,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (document_id) REFERENCES documents(document_id)
);

CREATE TABLE IF NOT EXISTS notifications (
    notification_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message VARCHAR(255),
    read BOOLEAN,
    user_id UUID,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);

CREATE TABLE IF NOT EXISTS otp_entries (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    otp_code VARCHAR(10) NOT NULL,
    pre_auth_token VARCHAR(64) NOT NULL,
    failed_attempts INT DEFAULT 0,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_otp_entries_email ON otp_entries(email);
CREATE INDEX idx_otp_entries_pre_auth_token ON otp_entries(pre_auth_token);