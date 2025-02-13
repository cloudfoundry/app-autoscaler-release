FROM jetpackio/devbox:latest@sha256:3e9b26747056d019dcedf295604a9b42dfe5a3b326fcc76de712b54a674314bd

# Installing your devbox project
WORKDIR /code
USER root:root
RUN mkdir -p /code && chown ${DEVBOX_USER}:${DEVBOX_USER} /code
USER ${DEVBOX_USER}:${DEVBOX_USER}
COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} devbox.json devbox.json
COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} devbox.lock devbox.lock


# Step 6: Copying local flakes directories
COPY local-flake local-flake
COPY local-flake local-flake

RUN devbox run -- echo "Installed Packages."

RUN devbox shellenv --init-hook >> ~/.profile
