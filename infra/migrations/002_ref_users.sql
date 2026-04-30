-- migrate:up
create table if not exists ref.users (
  user_id      text primary key,
  email        text unique not null,
  display_name text not null,
  created_at   timestamptz default now()
);

insert into ref.users (user_id, email, display_name)
values ('user_1', 'user@example.com', 'Single User')
on conflict (user_id) do nothing;

-- migrate:down
drop table if exists ref.users;
