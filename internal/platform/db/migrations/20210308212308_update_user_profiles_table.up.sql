ALTER TABLE user_profiles DROP COLUMN IF EXISTS certificate;
ALTER TABLE user_profiles ADD COLUMN department varchar(50) default null;
ALTER TABLE user_profiles ADD COLUMN date_severance date default null;
ALTER TABLE user_profiles ADD COLUMN reasons_severance varchar default null;