SELECT 'CREATE DATABASE postgres with owner postgres'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'postgres')\gexec;

create table  IF NOT EXISTS fib
(
    key integer,
    value integer
);

alter table fib owner to postgres;

create table  IF NOT EXISTS max
(
    max_fib_key integer,
    max_fib_value integer
);

alter table max owner to postgres;

INSERT INTO public.fib (key, value) VALUES (0, 0);
INSERT INTO public.fib (key, value) VALUES (1, 1);

INSERT INTO public.max (max_fib_key, max_fib_value) VALUES (1, 1);

