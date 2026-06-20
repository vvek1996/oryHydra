su - postgres
psql -U postgres

if failing: service postgresql start

CREATE DATABASE hydra;
CREATE DATABASE kratos;
CREATE USER ory WITH PASSWORD 'secret';
GRANT ALL PRIVILEGES ON DATABASE hydra TO ory;
GRANT ALL PRIVILEGES ON DATABASE kratos TO ory;


-- Give ownership of DB to ory
ALTER DATABASE hydra OWNER TO ory;
ALTER DATABASE kratos OWNER TO ory;

-- Connect to hydra DB
\c hydra

-- Give schema permissions
GRANT ALL ON SCHEMA public TO ory;

-- Optional but recommended:
ALTER SCHEMA public OWNER TO ory;

-- Also allow future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO ory;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO ory;




-- Make sure ory owns the database
ALTER DATABASE kratos OWNER TO ory;

-- Switch to kratos DB
\c kratos

-- Give schema access
GRANT ALL ON SCHEMA public TO ory;


-- Make ory the owner of schema
ALTER SCHEMA public OWNER TO ory;

-- Ensure future objects are allowed
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO ory;

ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO ory;


GRANT CREATE ON SCHEMA public TO ory;