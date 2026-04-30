-- migrate:up
create schema if not exists oms;
create schema if not exists portfolio;
create schema if not exists md;
create schema if not exists reports;
create schema if not exists ref;

-- migrate:down
drop schema if exists reports cascade;
drop schema if exists md cascade;
drop schema if exists portfolio cascade;
drop schema if exists oms cascade;
drop schema if exists ref cascade;
