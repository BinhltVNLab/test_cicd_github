create table if not exists user_roles(
	id serial primary key not null,
	name varchar(100) not null,
	description	text,
	created_at timestamp not null,
	updated_at timestamp not null,
	deleted_at timestamp
);

comment on column user_roles.id is 'Role id';
comment on column user_roles.deleted_at is 'Timestamp delete logic this record. When delete save current timestamp';
comment on column user_roles.name is 'Name';
comment on column user_roles.description is 'Decription';
comment on column user_roles.created_at is 'Save date when create';
comment on column user_roles.updated_at is 'Save date when update';

