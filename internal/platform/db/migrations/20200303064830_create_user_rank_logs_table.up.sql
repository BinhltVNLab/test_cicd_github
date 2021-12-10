create table if not exists user_rank_logs(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    user_id int not null,
    rank int not null
);

create index index_user_rank on user_rank_logs (user_id, rank);
create index index_user on user_rank_logs (user_id);

comment on column user_rank_logs.id is 'user_rank_log id';
comment on column user_rank_logs.created_at is 'Save timestamp when create';
comment on column user_rank_logs.updated_at is 'Save timestamp when update';
comment on column user_rank_logs.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column user_rank_logs.user_id is 'Id of user';
comment on column user_rank_logs.rank is 'Rank of user';
