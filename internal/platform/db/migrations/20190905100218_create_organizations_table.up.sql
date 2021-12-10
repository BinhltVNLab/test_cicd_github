create table if not exists organizations(
    id serial primary key not null,
    name varchar(100) not null,
    tag	varchar(50),
    phone_number varchar(20),
    address varchar(100),
    description text,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);
create index index_tag on organizations (tag);

comment on column organizations.id is 'Organization id';
comment on column organizations.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column organizations.name is 'Name';
comment on column organizations.tag is 'Tag is a unique string to indentify Organization. ';
comment on column organizations.phone_number is 'Phone number';
comment on column organizations.address is 'Address';
comment on column organizations.description is 'Description';
comment on column organizations.created_at is 'Save timestamp when create';
comment on column organizations.updated_at is 'Save timestamp when update';