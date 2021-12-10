create table if not exists registration_requests(
	id serial primary key not null,
	type smallint not null,
	status smallint not null,
	email varchar(100) not null,
	organization_id int not null,
	message text,
	created_at timestamp not null,
	updated_at timestamp not null,
	deleted_at timestamp
);

create index index_registration_requests_organization_id on registration_requests (organization_id);

comment on column registration_requests.id is 'id registration_request';
comment on column registration_requests.type is '_1: user send request admin to join organization _2: admin sent invite email';
comment on column registration_requests.status is 'Status request: _0: Request deny _1: Request pending _2: Request accept _3: Registered';
comment on column registration_requests.email is 'Email';
comment on column registration_requests.organization_id is 'Organization id. Foreign key to table organizations.id';
comment on column registration_requests.message is 'Message of user to request to join Organization';
comment on column registration_requests.created_at is 'Save date when create';
comment on column registration_requests.updated_at is 'Save date when update';
comment on column registration_requests.deleted_at is 'Delete logic record. When delete save current time';


