alter table target_evaluations alter column updated_by type integer;
alter table target_evaluations add column is_processing boolean;
