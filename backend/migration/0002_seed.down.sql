DELETE FROM standard WHERE standard_id IN (
  '563f2272-77c2-4445-883c-2a1c65fb82f8',
  '8464419c-d74e-4ec3-8ad3-ef4b163f3de5'
);

DELETE FROM services WHERE service_id IN (
  'ef61e242-1263-4899-a63f-346bc77c1ff3',
  '4c6aca5b-527e-4cae-bf0c-1b44c8ad8c43'
);

DELETE FROM document_types WHERE document_type_id IN (
  '9162b206-8947-48ab-9cac-26b4ba2ef153',
  '8e1e1800-dc00-4da4-b61d-8f925e9c2261',
  '4476092b-fc8c-4954-84cf-4578f4471ca1',
  '13d9d71b-4b3f-4577-b20b-f4510fcbec2d'
);

DELETE FROM users WHERE user_id IN (
  'c2a5ee05-e7ac-46b3-882f-24a8dedcb1b5',
  'bc6adda6-bfe6-4a6c-a637-35dd4ba4e562',
  'f95de891-4516-461c-b358-be2e4a442d9f'
);

DELETE FROM groups WHERE group_id IN (
  '9269058a-7cc4-4615-9a15-951d32fbbe88',
  '90b331d8-c334-474f-9b24-cc303ad8b537'
);