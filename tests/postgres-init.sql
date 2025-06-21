-- PostgreSQL initialization script for integration tests
-- This script increases connection limits and optimizes performance for concurrent testing

-- Increase max connections to handle concurrent integration tests
ALTER SYSTEM SET max_connections = 500;

-- Optimize shared memory settings
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';

-- Optimize maintenance settings
ALTER SYSTEM SET maintenance_work_mem = '64MB';

-- Optimize checkpoint settings
ALTER SYSTEM SET checkpoint_completion_target = 0.9;

-- Optimize WAL settings  
ALTER SYSTEM SET wal_buffers = '16MB';

-- Optimize query planner settings
ALTER SYSTEM SET default_statistics_target = 100;
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET effective_io_concurrency = 200;

-- Test-specific optimizations
-- Faster synchronous commit for tests (less durability, better performance)
ALTER SYSTEM SET synchronous_commit = 'off';

-- Reduce checkpoint frequency for tests
ALTER SYSTEM SET checkpoint_timeout = '15min';
ALTER SYSTEM SET checkpoint_warning = '0';

-- Optimize for concurrent connections
ALTER SYSTEM SET max_prepared_transactions = 100;
ALTER SYSTEM SET max_locks_per_transaction = 256;

-- Increase work memory for complex queries
ALTER SYSTEM SET work_mem = '32MB';

-- Log slow queries for debugging
ALTER SYSTEM SET log_min_duration_statement = '5000';

-- Apply the configuration changes
SELECT pg_reload_conf(); 