ALTER TABLE user_profiles ALTER COLUMN job_title TYPE integer USING (trim(job_title)::integer);
