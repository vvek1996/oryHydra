
mkdir -p /etc/kratos
nano /etc/kratos/kratos.yml

nano /etc/kratos/identity.schema.json

hydra create oauth2-client \
  --endpoint http://localhost:4445 \
  --name "test-client" \
  --grant-type authorization_code \
  --response-type code \
  --redirect-uri http://localhost:3000/callback \
  --scope openid \
  --scope offline \
  --secret secret


-------------------------------------- to delete client

hydra delete oauth2-client 9198f546-0bfd-4669-8891-e00d9a85b9ac \
  --endpoint http://localhost:4445

-------------------------------------- to get client id
hydra list oauth2-clients --endpoint http://localhost:4445

-------------------------------------- krato Registration flow
http://localhost:4433/self-service/registration/browser

