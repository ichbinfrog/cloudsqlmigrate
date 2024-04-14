# CloudSQL Migration utility

A tool for quickly migrating CloudSQL instances to each other. 

To run a migration, simply run:

```console
migrate \
    --src-project= \
    --src-instance= \
    --dst-project= \
    --dst-instance=S
```

Under the hood it heavily uses the [CloudSQL Admin API](https://cloud.google.com/sql/docs/mysql/admin-api) to: 

- Create an [on-demand snapshot](https://cloud.google.com/sql/docs/mysql/backup-recovery/backups#on-demand-backups) of the source instance 

- [Restore](https://cloud.google.com/sql/docs/mysql/backup-recovery/restoring) that snapshot in the target CloudSQL instance.


This approach is favorable the classic [sql export-import](https://cloud.google.com/sql/docs/mysql/import-export) flow for a couple reasons:

- It's much **faster** :zap:. The time to create and restore snapshots are quite well optimized is inversely proportional to the instance's resources.

- Users with their credentials, database schemas, permissions, extensions are preserved and transferred without issue from the source to the target instance.

- Not having to deal with the complexities of dumping and importing SQL (e.g. extension, user permission and other issues that arises from the lack of control of the dump generated)

- Does not stress the source instance if [automated backups](https://cloud.google.com/sql/docs/mysql/backup-recovery/backing-up#set-retention) is enabled. The latter is incremental thus generating a new snapshot from the last point in time is significantly faster and less resource intensive than performing an sql dump.


## Getting started
