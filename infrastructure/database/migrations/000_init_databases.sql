create schema if not exists core;

create table if not exists core.cars (
    id bigserial primary key,
    name varchar(1024) not null,
    colour char(7) not null, -- hex #RRGGBB
    price numeric(10, 2) not null,
    build_date date not null,
    created_at timestamp not null default timezone ('utc', now())
);