-- Delete default network in "queue_locks" table
DELETE FROM "queue_locks" WHERE "network_id" = 'default';
