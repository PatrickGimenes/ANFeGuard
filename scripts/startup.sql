--drop table portas;
CREATE TABLE portas (
    id SERIAL PRIMARY KEY,
    nome TEXT NOT NULL,
    porta_ws INT NOT NULL UNIQUE,
    porta_api INT NOT NULL UNIQUE,
    ambiente VARCHAR(10) NOT NULL
);

select * from portas;

--drop table servicos;
CREATE TABLE servicos (
    id SERIAL PRIMARY KEY,
    nome TEXT NOT null UNIQUE
);


--delete from servicos where id = 29;
select * from servicos;

--drop table service_logs;
CREATE TABLE service_logs (
    id SERIAL PRIMARY KEY,
    service_name TEXT NOT NULL,
    status TEXT NOT NULL,
    message TEXT,
    memory_percent NUMERIC,
    created_at TIMESTAMPTZ DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'America/Sao_Paulo')

);

select * from service_logs;

