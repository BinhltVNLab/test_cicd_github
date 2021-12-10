ALTER TABLE public.cvs DROP COLUMN media_id_other;
ALTER TABLE public.cvs ADD media_id_other VARCHAR(255) NULL;