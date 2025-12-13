create table if not exists users (
    id serial primary key,

        -- Basic information
        username varchar(50) not null unique,
        email varchar(100) not null unique,
        password_hash varchar(255) not null,
        phone_number varchar(15),

        -- Location information
        address text,
        city varchar(50),
        state varchar(50),
        zip_code varchar(10),
        latitude decimal(9,6),
        longitude decimal(9,6),

        -- Role and status
        role varchar(20) not null default 'customer',
        is_verified boolean default false,
        is_active boolean default true,

        -- Timestamps
        created_at timestamp default current_timestamp,
        updated_at timestamp default current_timestamp
    );

    -- Note: PostgreSQL doesn't support ON UPDATE CASCADE for timestamps
    -- Consider using a trigger or application-level updates for updated_at

    -- Indexes for improved query performance
    create index if not exists idx_users_email on users(email);
    create index if not exists idx_users_location on users(latitude, longitude);
    create index if not exists idx_users_is_active on users(is_active);
    create index if not exists idx_users_is_verified on users(is_verified);
    create index if not exists idx_users_city on users(city);
    create index if not exists idx_users_role on users(role);

    -- Comments
    comment on table users is 'User accounts table';
    comment on column users.role is 'Roles: customer, chef, admin';
    comment on column users.is_verified is 'Email verification status';
    comment on column users.is_active is 'Account active status';