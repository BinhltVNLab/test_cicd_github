create table if not exists asset_logs(
    id serial primary key not null,
    organization_id int not null,
    asset_id int not null,
    user_id int not null,
    start_day_using timestamp,
    end_day_using timestamp,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index index_asset_log_asset_id on asset_logs (asset_id);
create index index_asset_log_user_id on asset_logs (user_id);

comment on column asset_logs.id is 'id';
comment on column asset_logs.organization_id is 'organization id';
comment on column asset_logs.asset_id is 'asset id';
comment on column asset_logs.user_id is 'user id';
comment on column asset_logs.start_day_using is 'start day use asset';
comment on column asset_logs.end_day_using is 'end day use asset';
comment on column asset_logs.created_at is 'Save timestamp when create';
comment on column asset_logs.updated_at is 'Save timestamp when update';
comment on column asset_logs.deleted_at is 'Timestamp delete logic this record. When delete save current time';

alter table asset_logs add constraint asset_logs_asset_id foreign key (asset_id) references assets (id);
alter table asset_logs add constraint asset_logs_user_id foreign key (user_id) references users (id);