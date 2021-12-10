ALTER TABLE contract_types DROP COLUMN IF EXISTS file_template_name;
ALTER TABLE contract_types ADD COLUMN file_template_name varchar (150) unique;