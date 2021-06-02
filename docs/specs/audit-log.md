- [Audit Log](#audit-log)
  * [Storage Considerations](#storage-considerations)
    + [TimescaleDB](#timescaledb)
    + [PostgreSQL 10 native partitioning](#postgresql-10-native-partitioning)
    + [pg\_partman](#pg--partman)
  * [Fields](#fields)
    + [Core fields](#core-fields)
    + [Activity Type](#activity-type)
    + [IP address](#ip-address)
    + [User Agent](#user-agent)
    + [Client ID](#client-id)
  * [Database table schema](#database-table-schema)
  * [Admin API](#admin-api)
  * [Future Works](#future-works)

# Audit Log

Audit log records important user activities.

For the definition of events, see [Event](./event.md)

## Storage Considerations

To avoid further complicating the deployment requirements,
audit log is stored in a separate PostgreSQL database.
The PostgreSQL database can be a separate instance,
or it can be within the same instance, but a different database.
There exists a PostgreSQL extension called pg\_partman
which greatly simplify the management of partitioned tables.

Alternatively, audit log could be stored in Elasticsearch.
However, as of Elasticsearch 7, it does not support partitioning natively.
We need to create the index partition manually,
write specialized query with partition in mind,
and also drop indices that is not in the retention period.

### TimescaleDB

TimescaleDB is a time series database based on PostgreSQL.
It is available as managed service of self-hosted.
Self-hosting it seems too complicated.
So it is not preferred.

### PostgreSQL 10 native partitioning

Native partitioning is supported from PostgreSQL 10 onwards.
Since audit log is time series data,
partition can help reduce the size of the table and its indices.
Smaller tables and indices result in shorter insert time and query time.

### pg\_partman

[pg\_partman](https://github.com/pgpartman/pg_partman) is an extension to manage partition.

Major cloud providers such as Google, AWS and Azure provide this extension for their managed SQL database service.

- https://cloud.google.com/sql/docs/postgres/extensions#pg_partman
- https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/PostgreSQL_Partitions.html#PostgreSQL_Partitions.pg_partman
- https://docs.microsoft.com/en-us/azure/postgresql/concepts-extensions

We still have to call a stored procedure provided by pg\_partman to perform maintenance periodically.
That procedure creates new child tables and drop expired child tables.
For deployment on Kubernetes, we will use CronJob to run the procedure.

## Fields

### Core fields

Every log entry has `id`, `app_id`, `created_at`, `user_id` and `activity_type`.

### Activity Type

Activity Type is [the event type](./event.md#event-shape)

### IP address

`ip_address` is the IP address associated with the HTTP request when the log entry is created.

### User Agent

`user_agent` is the value of the HTTP User-Agent header when the log entry is created.

### Client ID

`client_id` is the client id associated with the event.

## Database table schema

```sql
CREATE TABLE _audit_log (
  -- Normally this should be PRIMARY KEY, but a partitioned table cannot have unique index on column that is not part of the partition key.
  id text NOT NULL,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  user_id text NOT NULL,
  activity_type text NOT NULL,
  ip_address inet,
  user_agent text,
  client_id text,
  data jsonb NOT NULL
) PARTITION BY RANGE (created_at);
CREATE INDEX _audit_log_idx_created_at_brin ON _audit_log USING BRIN (created_at);

CREATE TABLE _audit_log_template (LIKE _audit_log);
ALTER TABLE _audit_log_template ADD PRIMARY KEY (id);

SELECT create_parent(
  'schema._audit_log',
  'created_at',
  'native',
  'monthly',
  p_template_table := 'schema._audit_log_template'
);
UPDATE part_config
SET retention = '90 days', retention_keep_table = FALSE
WHERE parent_table = 'schema._audit_log';
```

## Admin API

The audit log is available in the Admin API.
Here is provisional preview of how it looks like in the GraphQL schema.

```graphql
type Query {
  # other root fields...

  auditLogs(
    # after, before, first and last are standard pagination arguments.
    after: String,
    before: String,
    first: Int,
    last: Int,
    # Must provide from and to fetch audit logs.
    rangeFrom: DateTime!,
    rangeTo: DateTime!,
    # By default logs of all activity types are returned.
    # Or only logs of selected activity types are returned.
    activityTypes: [AuditLogActivityType!]
  ): AuditLogConnection
}

type AuditLogConnection {
  edges: [AuditLogEdge]
  pageInfo: PageInfo!
  totalCount: Int
}

enum AuditLogActivityType {
  USER_AUTHENTICATED
  USER_SIGNED_OUT
}

type AuditLogEdge {
  cursor: String!
  node: AuditLog
}

type AuditLog implements Node {
  id: ID!
  createdAt: DateTime!
  activityType: AuditLogActivityType!
  user: User!
  ipAddress: String
  userAgent: String
  clientID: String
}
```

## Future Works

In the future, we may introduce activity types that start with `portal.`.
Those events are triggered by portal administrators.
The origin events and the portal events are likely to be stored together.
