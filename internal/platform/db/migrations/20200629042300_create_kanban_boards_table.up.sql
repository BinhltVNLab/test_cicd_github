create table if not exists kanban_boards(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    name varchar (255) not null,
    project_id int not null
);

create index index_kanban_boards_project_id on kanban_boards (project_id);

alter table kanban_boards add constraint fk_kanban_boards_project_id foreign key (project_id) references projects (id);

comment on column kanban_boards.id is 'kanban_board id';
comment on column kanban_boards.created_at is 'Save timestamp when create';
comment on column kanban_boards.updated_at is 'Save timestamp when update';
comment on column kanban_boards.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column kanban_boards.name is 'name of board';
comment on column kanban_boards.project_id is 'Id of project';
