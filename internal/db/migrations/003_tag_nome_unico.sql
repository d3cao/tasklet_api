-- Funde etiquetas de mesmo nome (sem diferenciar maiúsculas e espaços) na de menor id.
CREATE TEMP TABLE tag_duplicada ON COMMIT DROP AS
 SELECT tag_id, min(tag_id) OVER (PARTITION BY lower(btrim(nome))) AS manter FROM tag;
INSERT INTO possui_uma(task_id, tag_id)
 SELECT p.task_id, d.manter FROM possui_uma p JOIN tag_duplicada d USING(tag_id) WHERE d.tag_id <> d.manter
 ON CONFLICT DO NOTHING;
DELETE FROM tag WHERE tag_id IN (SELECT tag_id FROM tag_duplicada WHERE tag_id <> manter);
UPDATE tag SET nome = btrim(nome) WHERE nome <> btrim(nome);
CREATE UNIQUE INDEX tag_nome_unico ON tag (lower(nome));
