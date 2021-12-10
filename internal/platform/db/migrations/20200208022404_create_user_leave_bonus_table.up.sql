create table if not exists user_leave_bonus(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,

    organization_id int not null,
    user_id int not null,
    leave_bonus_type_id int not null,
    created_by int not null,
    updated_by int not null,
    year_belong int not null,
    reason varchar(100) not null,
    hour real not null
);

comment on column user_leave_bonus.id is 'user leave bonus id';
comment on column user_leave_bonus.created_at is 'Save timestamp when create';
comment on column user_leave_bonus.updated_at is 'Save timestamp when update';
comment on column user_leave_bonus.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column user_leave_bonus.organization_id is 'Id of organization';
comment on column user_leave_bonus.user_id is 'Id of user';
comment on column user_leave_bonus.leave_bonus_type_id is 'Id of leave bonus type';
comment on column user_leave_bonus.created_by is 'Created by user';
comment on column user_leave_bonus.updated_by is 'Updated by user';
comment on column user_leave_bonus.year_belong is 'Day off belong to year';
comment on column user_leave_bonus.reason is 'Reason';
comment on column user_leave_bonus.hour is 'Copy column hour of table user_leave_request';
