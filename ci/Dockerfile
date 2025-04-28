FROM jetpackio/devbox:latest

# Installing your devbox project
WORKDIR /code
USER root:root
RUN mkdir -p /code && chown ${DEVBOX_USER}:${DEVBOX_USER} /code
USER ${DEVBOX_USER}:${DEVBOX_USER}
COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} devbox.json devbox.json
COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} devbox.lock devbox.lock


# Step 6: Copying local flakes directories
COPY ./.devbox/virtenv/mysql/flake ./.devbox/virtenv/mysql/flake
COPY ./nix ./nix

RUN devbox run -- echo "Installed Packages."

# 🚧 To-do: Check if “ENTRYPOINT” is not the better alternative, see:
# <https://www.bmc.com/blogs/docker-cmd-vs-entrypoint>
CMD ["devbox", "shell"]
