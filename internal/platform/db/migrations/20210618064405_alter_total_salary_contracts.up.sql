ALTER TABLE public.contracts DROP COLUMN total_salary;
ALTER TABLE public.contracts ADD total_salary VARCHAR(50) NULL;