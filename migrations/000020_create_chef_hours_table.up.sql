-- Weekly working-hours windows per chef (#70). Times are minutes since
-- midnight in the platform time zone (Europe/Istanbul); opens_at > closes_at
-- means the window wraps past midnight (e.g. 18:00-02:00). A chef with NO
-- rows is always open — existing chefs keep behaving as before.
create table if not exists chef_hours (
    id serial primary key,
    chef_id integer not null references chefs(id) on delete cascade,

    weekday smallint not null check (weekday between 0 and 6), -- 0 = Sunday (Go time.Weekday)
    opens_at smallint not null check (opens_at between 0 and 1439),
    closes_at smallint not null check (closes_at between 0 and 1439),

    created_at timestamp not null default current_timestamp,

    unique (chef_id, weekday, opens_at)
);

create index if not exists idx_chef_hours_chef_id on chef_hours(chef_id);
