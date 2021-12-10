create table if not exists user_profiles(
	id serial primary key not null,
	user_id integer unique not null,
	avatar varchar(100),
	first_name varchar(100),
	last_name varchar(100),
	birthday date,
	created_at timestamp not null,
	updated_at timestamp not null,
	deleted_at timestamp
);

create index index_user_id on user_profiles (user_id);

comment on column user_profiles.id is 'profile id ';
comment on column user_profiles.user_id is 'User id, foreign key to table users.id';
comment on column user_profiles.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column user_profiles.avatar is 'Avatar';
comment on column user_profiles.first_name is 'First name';
comment on column user_profiles.last_name is 'Last name';
comment on column user_profiles.birthday is 'Birthday';
comment on column user_profiles.created_at is 'Save timestamp when create';
comment on column user_profiles.updated_at is 'Save timestamp when update';