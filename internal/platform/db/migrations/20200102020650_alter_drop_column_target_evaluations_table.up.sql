alter table target_evaluations alter column updated_by type integer USING (updated_by::integer);
alter table target_evaluations drop column is_processing;
