ALTER TABLE user_profiles ADD COLUMN current_address VARCHAR (255);
ALTER TABLE user_profiles ADD COLUMN tax_code VARCHAR (50) not null default 'tax_code';
ALTER TABLE user_profiles ADD COLUMN status int not null default 1;
ALTER TABLE user_profiles ADD COLUMN identity_card VARCHAR (50) UNIQUE;
ALTER TABLE user_profiles ADD COLUMN occupation VARCHAR (255);
ALTER TABLE user_profiles ADD COLUMN date_of_identity_card date;
ALTER TABLE user_profiles ADD COLUMN country VARCHAR (50);
ALTER TABLE user_profiles ADD COLUMN work_place VARCHAR (255);
ALTER TABLE user_profiles ADD COLUMN permanent_residence VARCHAR (255);

CREATE index index_profiles_status ON user_profiles (status);
