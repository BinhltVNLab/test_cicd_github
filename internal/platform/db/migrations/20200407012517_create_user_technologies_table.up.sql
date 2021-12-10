create table if not exists user_technologies(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    user_id int not null,
    technology_id int not null
);
create index index_user_technologies_user_id on user_technologies (user_id);
create index index_user_technologies_technology_id on user_technologies (technology_id);
create index index_user_technologies_user_technology on user_technologies (user_id, technology_id);

comment on column user_technologies.id is 'technology id';
comment on column user_technologies.created_at is 'Save timestamp when create';
comment on column user_technologies.updated_at is 'Save timestamp when update';
comment on column user_technologies.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column user_technologies.user_id is 'User id';
comment on column user_technologies.technology_id is 'Technology id';

alter table user_technologies add constraint user_technologies_users foreign key (user_id) references users (id);
alter table user_technologies add constraint user_technologies_technologies foreign key (technology_id) references technologies (id);
