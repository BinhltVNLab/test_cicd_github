alter table user_rank_logs add constraint user_rank_logs_user_id foreign key (user_id) references users (id);
