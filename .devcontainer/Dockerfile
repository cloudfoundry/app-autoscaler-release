FROM jetpackio/devbox:latest@sha256:3d164648f5ca0fa66cd080ee7c895af54e0a3ec19e19b232a2fea21628e82cfb

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
