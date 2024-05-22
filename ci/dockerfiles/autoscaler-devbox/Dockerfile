FROM jetpackio/devbox@sha256:430afb5a5c9b4ec9f56776e8fcb82fe67fa5fabba3ba67ebb1c13b17d0e8de8d


# Installing your devbox project
WORKDIR /code
USER root:root
RUN mkdir -p /code && chown "${DEVBOX_USER}:${DEVBOX_USER}" /code
USER ${DEVBOX_USER}:${DEVBOX_USER}
COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} devbox.json devbox.json
COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} devbox.lock devbox.lock


# Copying local flakes directories
COPY local-flake local-flake

RUN devbox run -- echo "Installed Packages."

CMD ["devbox", "shell"]
