export DSN=postgres://ory:secret@localhost:5432/hydra?sslmode=disable
export SECRETS_SYSTEM=some-secure-random-string

# export URLS_SELF_ISSUER=http://localhost:4444/
# export URLS_CONSENT=http://localhost:3000/consent
# export URLS_LOGIN=http://localhost:3000/login
# export URLS_LOGOUT=http://localhost:4000/logout

export URLS_SELF_ISSUER=http://localhost/.ory/hydra/
export URLS_LOGIN=http://localhost/login
export URLS_CONSENT=http://localhost/consent
export URLS_LOGOUT=http://localhost/api/logout


# hydra migrate sql -e --yes

hydra serve all --dev