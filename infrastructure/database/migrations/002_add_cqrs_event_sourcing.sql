create sequence if not exists core.car_id_seq;

select setval(
        'core.car_id_seq', greatest(coalesce(max(id), 1), 1), coalesce(max(id), 0) > 0
    )
from core.cars;

create table if not exists core.car_events (
    sequence_id bigserial primary key,
    aggregate_id bigint not null,
    event_type varchar(64) not null,
    payload jsonb not null,
    occurred_at timestamp not null default timezone ('utc', now()),
    constraint car_events_event_type_check check (
        event_type in (
            'car.created',
            'car.updated',
            'car.deleted'
        )
    )
);

create index if not exists car_events_aggregate_id_sequence_id_idx on core.car_events (aggregate_id, sequence_id);

create or replace function core.prevent_car_events_mutation()
returns trigger as $$
begin
    raise exception 'car_events is immutable';
end;
$$ language plpgsql;

drop trigger if exists prevent_car_events_update on core.car_events;

create trigger prevent_car_events_update
before update on core.car_events
for each row execute function core.prevent_car_events_mutation();

drop trigger if exists prevent_car_events_delete on core.car_events;

create trigger prevent_car_events_delete
before delete on core.car_events
for each row execute function core.prevent_car_events_mutation();

create table if not exists core.car_read_models (
    id bigint primary key,
    name varchar(1024) not null,
    colour char(7) not null,
    price numeric(10, 2) not null,
    build_date date not null,
    deleted_at timestamp null,
    projected_at timestamp not null default timezone ('utc', now())
);

insert into
    core.car_events (
        aggregate_id,
        event_type,
        payload,
        occurred_at
    )
select c.id, 'car.created', jsonb_build_object(
        'id', c.id, 'name', c.name, 'colour', c.colour, 'price', c.price, 'build_date', to_char(c.build_date, 'YYYY-MM-DD')
    ), c.created_at
from core.cars c
where
    not exists (
        select 1
        from core.car_events e
        where
            e.aggregate_id = c.id
            and e.event_type = 'car.created'
    );

truncate table core.car_read_models;

insert into
    core.car_read_models (
        id,
        name,
        colour,
        price,
        build_date
    )
select (payload ->> 'id')::bigint,
    payload ->> 'name',
    payload ->> 'colour',
    (payload ->> 'price')::numeric(10, 2),
    (payload ->> 'build_date')::date
from core.car_events
where
    event_type = 'car.created'
order by sequence_id;