--drop table portas;
CREATE TABLE portas (
    id SERIAL PRIMARY KEY,
    nome TEXT NOT NULL,
    porta_ws INT NOT NULL UNIQUE,
    porta_api INT NOT NULL UNIQUE,
    ambiente VARCHAR(10) NOT NULL
);
select *
from portas;
--drop table servicos;
CREATE TABLE servicos (
    id SERIAL PRIMARY KEY,
    nome VARCHAR(50) NOT NULL UNIQUE,
    displayName VARCHAR(50) NOT NULL UNIQUE,
    chave VARCHAR(80) NOT NULL UNIQUE,
    urlApi VARCHAR(80) NOT NULL,
    ativo SMALLINT NOT NULL DEFAULT 1
);
--delete from servicos where id = 3;
select *
from servicos;
--drop table service_logs;
CREATE TABLE service_logs (
    id SERIAL PRIMARY KEY,
    service_name TEXT NOT NULL,
    status TEXT NOT NULL,
    message TEXT,
    memory_percent NUMERIC,
    created_at TIMESTAMPTZ DEFAULT (
        CURRENT_TIMESTAMP AT TIME ZONE 'America/Sao_Paulo'
    )
);
select *
from service_logs;

--truncate table service_logs;
CREATE TABLE auditoria (
    id SERIAL PRIMARY KEY,
    service_name VARCHAR(200) NOT NULL,
    action VARCHAR(30) NOT NULL,
    username VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT (
        CURRENT_TIMESTAMP AT TIME ZONE 'America/Sao_Paulo'
    )
);