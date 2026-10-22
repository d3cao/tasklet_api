ALTER TABLE task
 ALTER COLUMN prazo TYPE TIMESTAMPTZ USING prazo AT TIME ZONE 'America/Recife',
 ALTER COLUMN dia_execucao TYPE TIMESTAMPTZ USING dia_execucao AT TIME ZONE 'America/Recife';
UPDATE task SET estado = 0 WHERE estado = 2;
ALTER TABLE task ADD COLUMN origem_id INTEGER UNIQUE REFERENCES task(task_id) ON DELETE SET NULL;
CREATE INDEX task_abertas_id ON task(task_id) WHERE estado <> 1;
CREATE INDEX possui_uma_tag_id ON possui_uma(tag_id, task_id);
