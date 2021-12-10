alter table user_leave_bonus add constraint user_leave_bonus_user_id foreign key (user_id) references users (id);
alter table user_leave_bonus add constraint user_leave_bonus_organization_id foreign key (organization_id) references organizations (id);
alter table user_leave_bonus add constraint user_leave_bonus_leave_bonus_type_id foreign key (leave_bonus_type_id) references leave_bonus_types (id);
