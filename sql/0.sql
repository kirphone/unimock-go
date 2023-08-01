create table templates
(
    id        INTEGER not null
        primary key autoincrement,
    name      TEXT    not null,
    body      TEXT    not null,
    subsystem TEXT
);

create unique index if not exists templates_name_uindex
    on templates (name);

create table if not exists triggers
(
    id          INTEGER           not null
        primary key autoincrement,
    type        TEXT              not null,
    expression  CHAR(700)         not null,
    description TEXT              not null,
    active      BOOLEAN default 0 not null,
    headers     CHAR(700),
    subsystem   TEXT
);

create table if not exists scenario_steps
(
    id           INTEGER not null
        constraint id
            primary key autoincrement,
    order_number integer not null,
    value        TEXT,
    trigger_id   INTEGER
        constraint scenario_steps_Triggers_id_fk
            references triggers
            on update cascade on delete cascade,
    step_type    TEXT    not null
);

create unique index if not exists Triggers_expression_header
    on triggers (expression, headers);

