------------------------------------------- user_roles ------------------------------------------------
INSERT INTO user_roles (id, name, description, created_at, updated_at) VALUES (1, 'admin', 'admin role', NOW(), NOW());
INSERT INTO user_roles (id, name, description, created_at, updated_at) VALUES (2, 'manager', 'manager role', NOW(), NOW());
INSERT INTO user_roles (id, name, description, created_at, updated_at) VALUES (3, 'user', 'user role', NOW(), NOW());
INSERT INTO user_roles (id, name, description, created_at, updated_at) VALUES (4, 'general manager', 'general manager role', NOW(), NOW());

-----------------------------------------------leave_request_types----------------------------------
INSERT INTO leave_request_types (id, name, description, created_at, updated_at) VALUES (1, 'Full day off', 'Full day off, equal 8 hour', '2020-02-07', '2020-02-07');
INSERT INTO leave_request_types (id, name, description, created_at, updated_at) VALUES (2, 'Morning off', 'Morning off, equal 4 hour', '2020-02-07', '2020-02-07');
INSERT INTO leave_request_types (id, name, description, created_at, updated_at) VALUES (3, 'Afternoon off', 'Afternoon, equal 4 hour', '2020-02-07', '2020-02-07');
INSERT INTO leave_request_types (id, name, description, created_at, updated_at) VALUES (4, 'Late for work', 'Late for work', '2020-02-07', '2020-02-07');
INSERT INTO leave_request_types (id, name, description, created_at, updated_at) VALUES (5, 'Leave early', 'Leave early', '2020-02-07', '2020-02-07');
INSERT INTO leave_request_types (id, name, description, created_at, updated_at) VALUES (6, 'Go outside', 'Go outside', '2020-02-07', '2020-02-07');
INSERT INTO leave_request_types (id, name, description, created_at, updated_at) VALUES (7, 'Work at home', 'Work at home', '2020-02-07', '2020-02-07');
INSERT INTO leave_request_types (id, name, description, created_at, updated_at) VALUES (8, 'Business trip', 'Business trip', '2020-02-07', '2020-02-07');
INSERT INTO leave_request_types (id, name, description, created_at, updated_at) VALUES (9, 'Other Leave', 'Other Leave', '2020-02-07', '2020-02-07');

----------------------------------------------leave_bonus_types---------------------------------------
INSERT INTO leave_bonus_types (id, name, description, created_at, updated_at) VALUES (1, 'Annual leave', 'Annual leave, add 12 days', '2020-02-07', '2020-02-07');
INSERT INTO leave_bonus_types (id, name, description, created_at, updated_at) VALUES (2, 'Seniority leave', 'Seniority, add (+x days)', '2020-02-07', '2020-02-07');
INSERT INTO leave_bonus_types (id, name, description, created_at, updated_at) VALUES (3, 'Sick leave', 'Sick leave, add (+x days)', '2020-02-07', '2020-02-07');
INSERT INTO leave_bonus_types (id, name, description, created_at, updated_at) VALUES (4, 'Marry leave', 'Marry, add (+x days)', '2020-02-07', '2020-02-07');
INSERT INTO leave_bonus_types (id, name, description, created_at, updated_at) VALUES (5, 'Maternity leave', 'Maternity leave, add (+x days)', '2020-02-07', '2020-02-07');
INSERT INTO leave_bonus_types (id, name, description, created_at, updated_at) VALUES (6, 'Bereavement leave', 'Bereavement (+ x day)', '2020-02-07', '2020-02-07');
INSERT INTO leave_bonus_types (id, name, description, created_at, updated_at) VALUES (7, 'Clear leave', 'Clear excess day off (- x day)', '2020-02-07', '2020-02-07');
INSERT INTO leave_bonus_types (id, name, description, created_at, updated_at) VALUES (8, 'Overtime leave', 'Overtime leave', '2020-02-07', '2020-02-07');
