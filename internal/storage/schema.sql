-- Первый шаг схемы. Новые изменения — отдельными шагами в db.go, этот файл после выпуска не меняется.

create table if not exists chat (
  id         integer primary key,
  started_at text not null default (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
