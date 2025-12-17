# pgm

Утилита для управления миграциями postgres sql

Пример использования:

```shell
# Создаст файлы миграции в директории migrations с названием {ts}_initial.{up/down}.sql.
pgm --command=create --migrationsDir=./migrations --migrationName=initial
```

```shell
# Применит все миграции из директории migrations к базе данных.
pgm |
  --command=migrate |
  --migrationsDir=./migrations |
  --migrationsTableSchema=jobs |
  --migrationsTable=migrations |
  --connectionString="..." |
  --priority=db
```

```shell
# Применит все миграции из директории migrations к базе данных.
# Если будут найдены несоответвия с файлами - то все применненные миграции к базе
# будут отменены до момента не соответсвия, а потом будут применены миграции из файлов.
pgm |
  --command=migrate |
  --migrationsDir=./migrations |
  --migrationsTableSchema=jobs |
  --migrationsTable=migrations |
  --connectionString="..." |
  --priority=fs
```
