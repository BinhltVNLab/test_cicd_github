create table if not exists assets(
    id serial primary key not null,
    organization_id int not null,
    user_id int,
    asset_type_id int not null,
    branch_id int not null,
    managed_by int not null,
    asset_name varchar (150) not null,
    asset_code varchar (150) unique not null,
    description varchar (300),
    date_of_purchase date,
    depreciation_period int,
    status SMALLINT not null,
    purchase_price bigint,
    license_end_date date,
    date_started_use date,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index index_assets_organization_id on assets (organization_id);
create index index_assets_user_id on assets (user_id);
create index index_assets_asset_type_id on assets (asset_type_id);
create index index_assets_branch_id on assets (organization_id);
create index index_assets_managed_by on assets (managed_by);

comment on column assets.id is 'id';
comment on column assets.user_id is 'user id';
comment on column assets.organization_id is 'organization id';
comment on column assets.asset_type_id is 'type id of asset';
comment on column assets.branch_id is 'branch id';
comment on column assets.managed_by is 'asset managed by';
comment on column assets.asset_name is 'asset name';
comment on column assets.description is 'description';
comment on column assets.date_of_purchase is 'date of purchase asset';
comment on column assets.depreciation_period is 'depreciation period asset';
comment on column assets.status is 'status';
comment on column assets.purchase_price is 'purchase price of asset';
comment on column assets.license_end_date is 'license end date';
comment on column assets.date_started_use is 'date started use';
comment on column assets.asset_code is 'asset code';
comment on column assets.created_at is 'Save timestamp when create';
comment on column assets.updated_at is 'Save timestamp when update';
comment on column assets.deleted_at is 'Timestamp delete logic this record. When delete save current time';

alter table assets add constraint asset_organization_id foreign key (organization_id) references organizations (id);
alter table assets add constraint asset_user_id foreign key (user_id) references user_profiles (user_id);
alter table assets add constraint asset_asset_type_id foreign key (asset_type_id) references asset_types (id);
alter table assets add constraint asset_branch_id foreign key (branch_id) references branches (id);
alter table assets add constraint asset_managed_by foreign key (managed_by) references user_profiles (user_id);