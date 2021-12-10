alter table registration_requests add constraint registration_requests_organization_id foreign key (organization_id) references organizations (id);

alter table registration_codes add constraint registration_codes_registration_request_id foreign key (registration_request_id) references registration_requests (id);
