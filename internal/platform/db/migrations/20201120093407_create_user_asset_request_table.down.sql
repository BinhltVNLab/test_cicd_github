alter table user_asset_requests drop constraint if exists asset_request_asset_id;
alter table user_asset_requests drop constraint if exists asset_request_created_by;
alter table user_asset_requests drop constraint if exists asset_request_organization_id;
drop table if exists user_asset_requests;
