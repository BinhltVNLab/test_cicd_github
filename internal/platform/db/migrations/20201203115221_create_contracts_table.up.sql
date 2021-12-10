create table if not exists contracts(
    id serial primary key not null,
    user_id int not null,
    organization_id int not null,
    contract_type_id int not null,
    insurance_salary bigint not null,
    total_salary bigint,
    contract_start_date date not null,
    contract_end_date date,
    currency_unit int,
    file_name varchar (255) not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp
);

create index index_contracts_user_id on contracts (user_id);
create index index_contracts_organization_id on contracts (organization_id);
create index index_contracts_contract_type_id on contracts (contract_type_id);
create index index_contracts_insurance_salary on contracts (insurance_salary);
create index index_contracts_total_salary on contracts (total_salary);

comment on column contracts.id is 'id';
comment on column contracts.user_id is 'user id';
comment on column contracts.organization_id is 'organization id';
comment on column contracts.contract_type_id is 'type id of contract';
comment on column contracts.insurance_salary is 'insurance salary of contract';
comment on column contracts.total_salary is 'tatal of salary';
comment on column contracts.contract_start_date is 'date started contract';
comment on column contracts.contract_end_date is 'date end contract';
comment on column contracts.currency_unit is 'currency unit';
comment on column contracts.file_name is 'file name of contract';
comment on column contracts.created_at is 'Save timestamp when create';
comment on column contracts.updated_at is 'Save timestamp when update';
comment on column contracts.deleted_at is 'Timestamp delete logic this record. When delete save current time';

alter table contracts add constraint contracts_contract_type_id foreign key (contract_type_id) references contract_types (id);
alter table contracts add constraint contracts_user_id foreign key (user_id) references user_profiles (user_id);
alter table contracts add constraint contracts_organization_id foreign key (organization_id) references organizations (id);
