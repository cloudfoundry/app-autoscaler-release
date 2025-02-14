FROM jetpackio/devbox:latest@sha256:fb517afe634cf0b176e930441892cc1b2a97aa95636acb98db4f760c0a68a517

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
