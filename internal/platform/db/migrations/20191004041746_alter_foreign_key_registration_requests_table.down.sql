alter table registration_requests drop constraint if exists registration_requests_organization_id;

alter table registration_codes drop constraint if exists registration_codes_registration_request_id;
