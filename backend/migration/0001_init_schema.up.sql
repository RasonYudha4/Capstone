CREATE TYPE user_role AS ENUM ('master-admin', 'admin', 'staff');
CREATE TYPE audit_type AS ENUM ('insert', 'open', 'update','delete','error');
CREATE TYPE audit_source AS ENUM ('client', 'system');

CREATE TABLE IF NOT EXISTS groups (
    group_id UUID PRIMARY KEY ,
    group_name VARCHAR(255),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY,
    group_id UUID,
    email VARCHAR(255),
    verified BOOLEAN,
    role user_role,
    verification_code VARCHAR(6),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(group_id)
);

CREATE TABLE IF NOT EXISTS services (
    service_id UUID PRIMARY KEY,
    group_id UUID,
    service_code VARCHAR(20),
    description VARCHAR(255),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(group_id)
);

CREATE TABLE IF NOT EXISTS standard (
    standard_id UUID PRIMARY KEY,
    service_id UUID,
    standard_code VARCHAR(20),
    description VARCHAR(255),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (service_id) REFERENCES services(service_id)
);

CREATE TABLE IF NOT EXISTS assessment (
    assessment_id UUID PRIMARY KEY,
    standard_id UUID,
    assessment_code VARCHAR(20),
    description VARCHAR(255),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (standard_id) REFERENCES standard(standard_id)
);

CREATE TABLE IF NOT EXISTS document_types (
    document_type_id UUID PRIMARY KEY,
    name VARCHAR(255),
    description VARCHAR(255),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS documents (
    document_id UUID PRIMARY KEY ,
    assessment_id UUID,
    filename VARCHAR(255),
    filepath VARCHAR(255),
    document_type_id UUID,
    status VARCHAR(20),
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    created_by UUID,
    group_id UUID,
    service_id UUID,
    standard_id UUID,
    FOREIGN KEY (assessment_id) REFERENCES assessment(assessment_id),
    FOREIGN KEY (document_type_id) REFERENCES document_types(document_type_id),
    FOREIGN KEY (created_by) REFERENCES users(user_id),
    FOREIGN KEY (group_id) REFERENCES groups(group_id),
    FOREIGN KEY (service_id) REFERENCES services(service_id),
    FOREIGN KEY (standard_id) REFERENCES standard(standard_id)
);

CREATE TABLE IF NOT EXISTS audit (
    audit_id UUID PRIMARY KEY,
    type audit_type,
    action VARCHAR(20),
    user_id UUID,
    document_id UUID,
    source audit_source,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (document_id) REFERENCES documents(document_id)
);

CREATE TABLE IF NOT EXISTS notifications (
    notification_id UUID PRIMARY KEY,
    message VARCHAR(255),
    read BOOLEAN,
    user_id UUID,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);