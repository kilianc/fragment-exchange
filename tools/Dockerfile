# tools/Dockerfile — the project's pinned toolchain.
#
# Nothing fx builds or ships needs Node. This image exists for one reason: the
# Vercel CLI is a Node program, and it is going to stay in a container rather
# than on the host, where an install script from any transitive dependency
# would run with the privileges of whoever typed the command.
#
# Versions are pinned. This file is the lockfile.
#
#   make tools        # build it
#   bin/vercel --help # use it
FROM node:22.12.0-bookworm-slim

RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates git \
 && rm -rf /var/lib/apt/lists/*

RUN npm install -g --ignore-scripts vercel@58.9.4

ENV NPM_CONFIG_UPDATE_NOTIFIER=false \
    NPM_CONFIG_FUND=false

WORKDIR /work
