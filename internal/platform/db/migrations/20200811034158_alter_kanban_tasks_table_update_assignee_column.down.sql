ALTER TABLE kanban_tasks RENAME assignees TO assignee;
ALTER TABLE kanban_tasks ALTER COLUMN assignee TYPE INTEGER,
alter table kanban_tasks add constraint fk_kanban_tasks_assignee foreign key (assignee) references users (id);
