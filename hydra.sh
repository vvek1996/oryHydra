export DSN=postgres://ory:secret@localhost:5432/hydra?sslmode=disable
export SECRETS_SYSTEM=some-secure-random-string

# export URLS_SELF_ISSUER=http://localhost:4444/
# export URLS_CONSENT=http://localhost:3000/consent
# export URLS_LOGIN=http://localhost:3000/login
# export URLS_LOGOUT=http://localhost:4000/logout

export URLS_SELF_ISSUER=http://localhost:8080/.ory/hydra/
export URLS_LOGIN=http://localhost:8080/login
export URLS_CONSENT=http://localhost:8080/consent
export URLS_LOGOUT=http://localhost:8080/api/logout


# hydra migrate sql -e --yes

hydra serve all --dev