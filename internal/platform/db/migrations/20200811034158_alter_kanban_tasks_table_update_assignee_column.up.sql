ALTER TABLE kanban_tasks DROP CONSTRAINT fk_kanban_tasks_assignee;
ALTER TABLE kanban_tasks RENAME assignee TO assignees;
ALTER TABLE kanban_tasks ALTER COLUMN assignees TYPE INTEGER [] USING array[assignees]::integer[];
