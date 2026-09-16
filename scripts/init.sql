CREATE TABLE task (
    task_id SERIAL PRIMARY KEY,
    nome VARCHAR(50) NOT NULL,
    descricao TEXT,
    estado INTEGER NOT NULL DEFAULT 0 CHECK (estado IN (0, 1, 2)),
    prazo TIMESTAMP,
    repeticao INTEGER NOT NULL DEFAULT 0 CHECK (repeticao >= 0),
    dia_execucao TIMESTAMP
);

CREATE TABLE tag (
    tag_id SERIAL PRIMARY KEY,
    nome VARCHAR(50) NOT NULL
);

CREATE TABLE possui_uma (
    task_id INTEGER,
    tag_id INTEGER,
    PRIMARY KEY (task_id, tag_id),
    FOREIGN KEY (task_id) REFERENCES task(task_id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tag(tag_id) ON DELETE CASCADE
);