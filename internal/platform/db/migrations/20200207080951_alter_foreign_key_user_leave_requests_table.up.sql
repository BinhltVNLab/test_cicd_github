alter table user_leave_requests add constraint user_leave_requests_user_id foreign key (user_id) references users (id);
alter table user_leave_requests add constraint user_leave_requests_organization_id foreign key (organization_id) references organizations (id);
alter table user_leave_requests add constraint user_leave_requests_leave_request_type_id foreign key (leave_request_type_id) references leave_request_types (id);
