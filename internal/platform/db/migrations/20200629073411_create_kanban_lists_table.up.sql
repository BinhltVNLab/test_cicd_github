create table if not exists kanban_lists(
    id serial primary key not null,
    created_at timestamp not null,
    updated_at timestamp not null,
    deleted_at timestamp,
    name varchar (255) not null,
    kanban_board_id int not null,
    position_in_board int not null
);

create index index_kanban_lists_kanban_board_id on kanban_lists (kanban_board_id);

alter table kanban_lists add constraint fk_kanban_lists_kanban_board_id foreign key (kanban_board_id) references kanban_boards (id);

comment on column kanban_lists.id is 'list id';
comment on column kanban_lists.created_at is 'Save timestamp when create';
comment on column kanban_lists.updated_at is 'Save timestamp when update';
comment on column kanban_lists.deleted_at is 'Timestamp delete logic this record. When delete save current time';
comment on column kanban_lists.name is 'name of column';
comment on column kanban_lists.kanban_board_id is 'Id of board';
comment on column kanban_lists.position_in_board is 'position in board';
