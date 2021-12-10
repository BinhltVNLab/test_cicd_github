create table if not exists user_asset_requests(
    id serial primary key not null,
    organization_id int not null,
    asset_id int not null,
    created_by int not null,
    status int,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index index_asset_request_asset_id on user_asset_requests (asset_id);
create index index_asset_request_created_by on user_asset_requests (created_by);
create index index_asset_request_organization_id on user_asset_requests (organization_id);
create index index_asset_request_status on user_asset_requests (status);

comment on column user_asset_requests.id is 'id';
comment on column user_asset_requests.organization_id is 'organization id';
comment on column user_asset_requests.asset_id is 'asset id';
comment on column user_asset_requests.created_by is 'created_by id';
comment on column user_asset_requests.status is 'status';
comment on column user_asset_requests.created_at is 'Save timestamp when create';
comment on column user_asset_requests.updated_at is 'Save timestamp when update';
comment on column user_asset_requests.deleted_at is 'Timestamp delete logic this record. When delete save current time';

alter table user_asset_requests add constraint asset_request_organization_id foreign key (organization_id) references organizations (id);
alter table user_asset_requests add constraint asset_request_asset_id foreign key (asset_id) references assets (id);
alter table user_asset_requests add constraint asset_request_created_by foreign key (created_by) references user_profiles (id);
