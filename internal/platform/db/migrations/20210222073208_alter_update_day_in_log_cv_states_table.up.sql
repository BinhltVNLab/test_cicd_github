alter table log_cv_states alter column update_day type timestamp(6)
    without time zone using update_day::timestamp(6) without time zone;