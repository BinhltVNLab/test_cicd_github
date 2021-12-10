create table if not exists users(
    id serial primary key not null,
    organization_id integer,
    email varchar(100) not null,
    password varchar(100),
    role_id integer not null,
    phone_number varchar(20),
    last_login_time timestamp not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index index_email on users (email);
create index index_organizationid on users (organization_id);
create index index_email_organizationid on users (email, organization_id);

comment on column users.id  is 'User id';
comment on column users.deleted_at is 'Delete logic record. When delete save current time';
comment on column users.organization_id is 'Organization id. Foreign key to table organizations.id';
comment on column users.email  is 'Email';
comment on column users.password is 'Password hash Md5 If user create by facebook, gmail..., password will be empty';
comment on column users.role_id is 'Role of user, foreign key to table user_roles.id';
comment on column users.phone_number is 'Phone number';
comment on column users.last_login_time is 'Save timestamp when user login';
comment on column users.created_at is 'Save timestamp when create';
comment on column users.updated_at is 'Save timestamp when update';