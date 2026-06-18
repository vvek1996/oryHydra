export KRATOS_PUBLIC_CORS_ENABLED=true
export KRATOS_PUBLIC_CORS_ALLOWED_ORIGINS=http://localhost:3000

export SERVE_PUBLIC_CORS_ENABLED=true
export SERVE_PUBLIC_CORS_ALLOWED_ORIGINS=http://localhost:3000

# kratos migrate sql -e -c /etc/kratos/kratos.yml -y
kratos serve -c /etc/kratos/kratos.yml --dev