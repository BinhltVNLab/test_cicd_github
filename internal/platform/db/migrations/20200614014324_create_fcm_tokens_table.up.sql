create table if not exists fcm_tokens(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    user_id integer not null,
    token varchar(255) not null
);

create index index_fcm_tokens_user_id on fcm_tokens (user_id);

alter table fcm_tokens add constraint fk_fcm_tokens_user_id foreign key (user_id) references users (id);

comment on column fcm_tokens.id is 'firebase_token id';
comment on column fcm_tokens.created_at is 'Save timestamp when create';
comment on column fcm_tokens.updated_at is 'Save timestamp when update';
comment on column fcm_tokens.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column fcm_tokens.user_id is 'Id of user';
comment on column fcm_tokens.token is 'token of user';
