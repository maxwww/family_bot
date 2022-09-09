CREATE TABLE users
(
    id            serial       not null unique,
    telegram_id   integer      not null unique,
    first_name    varchar(255) not null,
    last_name     varchar(255),
    user_name     varchar(255)
);

CREATE TABLE tasks
(
    id    serial       not null unique,
    title varchar(255) not null,
    date  timestamp,
    done  boolean      not null default false
);