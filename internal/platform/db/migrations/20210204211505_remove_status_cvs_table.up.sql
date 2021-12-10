alter table cvs drop column if exists status;
alter table cvs alter column media_id drop not null;