create table if not exists kanban_tasks(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    kanban_list_id int not null,
    title varchar(255) not null,
    description varchar(510),
    due_date timestamp,
    assignee int,
    status smallint not null,
    position_in_list int not null,
    created_by int not null,
    updated_by int not null
);

create index index_kanban_tasks_title on kanban_tasks (title);
create index index_kanban_tasks_assignee on kanban_tasks (assignee);
create index index_kanban_tasks_status on kanban_tasks (status);

alter table kanban_tasks add constraint fk_kanban_tasks_kanban_list_id foreign key (kanban_list_id) references kanban_lists (id);
alter table kanban_tasks add constraint fk_kanban_tasks_assignee foreign key (assignee) references users (id);

comment on column kanban_tasks.id is 'kanban_task id';
comment on column kanban_tasks.created_at is 'Save timestamp when create';
comment on column kanban_tasks.updated_at is 'Save timestamp when update';
comment on column kanban_tasks.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column kanban_tasks.kanban_list_id is 'Id of list';
comment on column kanban_tasks.title is 'Title of task';
comment on column kanban_tasks.description is 'Content of task';
comment on column kanban_tasks.due_date is 'Due date task';
comment on column kanban_tasks.assignee is 'member is assigned for task';
comment on column kanban_tasks.status is 'status of task';
comment on column kanban_tasks.position_in_list is 'position of task';
comment on column kanban_tasks.created_by is 'Created by';
comment on column kanban_tasks.updated_by is 'Updated by';
