-- +migrate Up notransaction

-- _portal_domain has no indexes on `domain` or `app_id`. Every request
-- resolves the tenant via GetAppIDByDomain (WHERE domain = ?) and then
-- loads all domains via GetDomainsByAppID (WHERE app_id = ?). Both do
-- full table scans. At cloud scale the table grows to tens of thousands
-- of rows and cache-miss resolution becomes measurably slow.
--
-- CONCURRENTLY builds the index without taking a ShareLock, so reads and
-- writes continue uninterrupted during the build. notransaction above is
-- required because CONCURRENTLY cannot run inside a transaction.
CREATE INDEX CONCURRENTLY _portal_domain_domain ON _portal_domain (domain);
CREATE INDEX CONCURRENTLY _portal_domain_app_id ON _portal_domain (app_id);

-- +migrate Down notransaction
DROP INDEX CONCURRENTLY _portal_domain_domain;
DROP INDEX CONCURRENTLY _portal_domain_app_id;
