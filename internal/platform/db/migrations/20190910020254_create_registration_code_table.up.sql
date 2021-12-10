create table if not exists registration_codes(
	id serial primary key not null,
	email varchar(100) not null,
	code text,
	registration_request_id int,
	created_at timestamp not null,
	updated_at timestamp not null,
	expired_at timestamp not null,
	deleted_at timestamp
);

create index index_registration_codes_email on registration_codes (email);
create index index_registration_codes_code on registration_codes (code);

comment on column registration_codes.id is 'registration_code id';
comment on column registration_codes.email is 'register email for create code';
comment on column registration_codes.code is 'generated code for registration';
comment on column registration_codes.registration_request_id is 'requestid from registration_requests';
comment on column registration_codes.created_at is 'Save date when create';
comment on column registration_codes.updated_at is 'Save date when update';
comment on column registration_codes.expired_at is 'expired date for code';
comment on column registration_codes.deleted_at is 'deleted_at date for code';

