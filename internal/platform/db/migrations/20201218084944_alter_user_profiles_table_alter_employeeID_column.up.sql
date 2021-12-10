ALTER TABLE user_profiles DROP COLUMN IF EXISTS employee_id;
ALTER TABLE user_profiles ADD COLUMN employee_id varchar (20) unique;