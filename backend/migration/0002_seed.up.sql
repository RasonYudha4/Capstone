INSERT INTO groups (group_id, group_name, created_at, updated_at) VALUES
  ('9269058a-7cc4-4615-9a15-951d32fbbe88', 'Head Office', NOW(), NOW()),
  ('90b331d8-c334-474f-9b24-cc303ad8b537', 'Branch A',    NOW(), NOW());

INSERT INTO users (user_id, group_id, email, verified, role, verification_code, created_at, updated_at) VALUES
  ('c2a5ee05-e7ac-46b3-882f-24a8dedcb1b5', '9269058a-7cc4-4615-9a15-951d32fbbe88', 'masteradmin@gmail.com', true, 'master-admin', '000000', NOW(), NOW()),
  ('bc6adda6-bfe6-4a6c-a637-35dd4ba4e562', '9269058a-7cc4-4615-9a15-951d32fbbe88', 'admin@gmail.com',       true, 'admin',        '000000', NOW(), NOW()),
  ('f95de891-4516-461c-b358-be2e4a442d9f', '90b331d8-c334-474f-9b24-cc303ad8b537', 'staff@gmail.com',       true, 'staff',        '000000', NOW(), NOW());

INSERT INTO document_types (document_type_id, name, description, created_at, updated_at) VALUES
  ('9162b206-8947-48ab-9cac-26b4ba2ef153', 'Contract',    'Legal contract documents', NOW(), NOW()),
  ('8e1e1800-dc00-4da4-b61d-8f925e9c2261', 'Invoice',     'Payment invoice documents', NOW(), NOW()),
  ('4476092b-fc8c-4954-84cf-4578f4471ca1', 'Report',      'General report documents',  NOW(), NOW()),
  ('13d9d71b-4b3f-4577-b20b-f4510fcbec2d', 'Certificate', 'Certification documents',   NOW(), NOW());

INSERT INTO services (service_id, group_id, service_code, description, created_at, updated_at) VALUES
  ('4c6aca5b-527e-4cae-bf0c-1b44c8ad8c43', '9269058a-7cc4-4615-9a15-951d32fbbe88', 'SVC-001', 'Document Management',  NOW(), NOW()),
  ('ef61e242-1263-4899-a63f-346bc77c1ff3', '9269058a-7cc4-4615-9a15-951d32fbbe88', 'SVC-002', 'Approval Services',  NOW(), NOW());

INSERT INTO standard (standard_id, service_id, standard_code, description, created_at, updated_at) VALUES
  ('563f2272-77c2-4445-883c-2a1c65fb82f8', '4c6aca5b-527e-4cae-bf0c-1b44c8ad8c43', 'STD-001', 'ISO 9001',    NOW(), NOW()),
  ('8464419c-d74e-4ec3-8ad3-ef4b163f3de5', 'ef61e242-1263-4899-a63f-346bc77c1ff3', 'STD-002', 'ISO 27001',   NOW(), NOW());

INSERT INTO assessment (assessment_id, standard_id, assessment_code, description, created_at, updated_at) VALUES
  ('90b331d8-c334-474f-9b24-cc303ad8b537', '563f2272-77c2-4445-883c-2a1c65fb82f8', 'ASM-001', 'Initial Assessment', NOW(), NOW()),
  ('d1e5e7c4-3a9b-4e5a-9c3f-2b1a6e7f8c9d', '8464419c-d74e-4ec3-8ad3-ef4b163f3de5', 'ASM-002', 'Risk Assessment',    NOW(), NOW());