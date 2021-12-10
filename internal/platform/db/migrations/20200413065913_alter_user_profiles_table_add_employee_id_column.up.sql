ALTER TABLE user_profiles ADD COLUMN employee_id VARCHAR(10);
CREATE index index_employee_id ON user_profiles (employee_id);
