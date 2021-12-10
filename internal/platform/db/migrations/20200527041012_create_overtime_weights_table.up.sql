create table if not exists overtime_weights(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    organization_id integer not null,
    normal_day_weight real,
    weekend_weight real,
    holiday_weight real
);

alter table overtime_weights add constraint ot_weight_organization_id foreign key (organization_id) references organizations (id);

comment on column overtime_weights.id is 'overtime_weights id';
comment on column overtime_weights.created_at is 'Save timestamp when create';
comment on column overtime_weights.updated_at is 'Save timestamp when update';
comment on column overtime_weights.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column overtime_weights.organization_id is 'organization id';
comment on column overtime_weights.normal_day_weight is 'Weight of normal day';
comment on column overtime_weights.weekend_weight is 'Weight of weekend';
comment on column overtime_weights.holiday_weight is 'Weight of holiday';
