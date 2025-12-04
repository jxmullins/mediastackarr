# TODO

## ISSUES
- `base-working-files/secure_authentik_database.sh` and `base-working-files/create_guacamole_database.sh` use `$ADMIN_PASSWORD` for `PGPASSWORD`, but that variable is never populated from `.env`; the psql grant blocks fail even when credentials are correct. Swap to `$POSTGRESQL_PASSWORD` or require input.
- `base-working-files/restart.sh` references an undefined `containers` variable before attempting `docker stop`/`docker rm`, so running containers for this stack never get stopped cleanly; the subsequent global `docker container prune` can nuke unrelated containers instead of only the stack.
- Traefik dashboard is published directly on `${WEBUI_PORT_TRAEFIK}:8080` while `traefik-static.yaml` sets `api.insecure: true`; host-level access to :8080 bypasses the configured forward-auth middleware and exposes the dashboard unauthenticated.

## GAPS
- No checked-in `.env.example` documenting required variables (PUID/PGID, dozens of port mappings, Cloudflare token, VPN credentials, etc.), making `docker compose config` fail by default and inviting misconfiguration.
- Healthchecks are defined only for a few services (postgresql, valkey, gluetun); most app containers lack them, so `depends_on: condition: service_healthy` can’t gate startup and failure recovery is weaker.
- The three compose variants duplicate large sections verbatim; there’s no shared base/override pattern or profiles to minimise drift between full/mini/no-VPN setups.

## IMPROVEMENT SUGGESTIONS
- Pin all images to specific versions or digests instead of `latest` (postgres, traefik, gluetun, linuxserver apps, etc.) to avoid surprise upgrades and to support reproducible restores.
- Remove redundant host port publishes for UI services that are already fronted by Traefik (authentik, grafana, guacamole, bazarr, etc.), or at least bind them to localhost; rely on the reverse proxy + SSO instead of exposing raw UIs.
- Externalise secrets to `.env` or Docker secrets (e.g., CrowdSec LAPI key, cookie secret, database passwords) and rotate the checked-in CrowdSec key in `base-working-files/traefik-dynamic.yaml`.
- Harden Traefik: disable `api.insecure`, gate metrics (`:8082`) and CrowdSec ports (6060/7422) to `127.0.0.1` or internal network, and enable the commented `contentSecurityPolicy` header once rules are tested.
- Add a thin lint/test layer (shellcheck for scripts, `docker compose config` in CI) to catch undefined variables and missing dependencies (`yq`, `xmllint`) early.

## SECURITY CONCERNS
- CrowdSec LAPI key is hard-coded in `base-working-files/traefik-dynamic.yaml`; anyone with repo access can abuse/ban-list the instance—move to secret storage and rotate immediately.
- Valkey/Redis is exposed on `${VALKEY_PORT}:6379` with no password and no network restriction, inviting unauthenticated access; limit binding or require `requirepass`.
- Multiple services expose host ports that bypass Traefik’s authentication (e.g., grafana on `${WEBUI_PORT_GRAFANA}:3000`, headscale on `${CONNECT_PORT_HEADSCALE}:8080`, guacamole on `${WEBUI_PORT_GUACAMOLE}:8080`, and all *arr/plex UIs via the gluetun container). These should be firewalled or removed in favour of the reverse proxy.
- Gluetun publishes many application ports directly on the host while those apps also sit behind Traefik; the direct bindings circumvent SSO and rate-limiting and could leak real IP if VPN drops—prefer proxy-only exposure with proper middleware.
- `base-working-files/get-apikeys.sh` prints *arr API keys in plaintext to stdout; if run on a multi-user box or logged centrally it leaks credentials—add a warning, restrict permissions, or write to a protected file.
