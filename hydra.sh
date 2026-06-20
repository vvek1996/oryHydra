export DSN=postgres://ory:secret@localhost:5432/hydra?sslmode=disable
export URLS_SELF_ISSUER=http://localhost:4444/
export URLS_CONSENT=http://localhost:3000/consent
export URLS_LOGIN=http://localhost:3000/login
export URLS_LOGOUT=http://localhost:4000/logout
export SECRETS_SYSTEM=some-secure-random-string

hydra serve all --dev