FROM jetpackio/devbox:latest

# Installing your devbox project
WORKDIR /code
USER root:root
RUN mkdir --parents /code && chown ${DEVBOX_USER}:${DEVBOX_USER} /code
USER ${DEVBOX_USER}:${DEVBOX_USER}
COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} devbox.json devbox.json
COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} devbox.lock devbox.lock


# # 💡 Here maybe the chown is missing!
# # Step 6: Copying local flakes directories
#COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} ./.devbox/virtenv/mysql/flake ./.devbox/virtenv/mysql/flake
COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} ./.devbox ./.devbox
COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} ./nix ./nix

# 💡 The issue probably is, that we don't copy the necessary files before.
# I alternatively will try to avoid the copies and read-only-mount those directories
# into the image.
RUN \
	# --mount=type=bind,source=./nix,target=/code/nix,readonly \
	# --mount=type=bind,source=./.devbox,target=/code/.devbox,readonly \
	<<-EODI
		echo "ls -lah /code: $(ls -lah /code)"
		echo "ls -lah /code/.devbox: $(ls -lah /code/.devbox)"
		echo "ls -lah /code/.devbox/gen: $(ls -lah /code/.devbox/gen)"
		devbox run -- echo "Installed Packages."
EODI

# Make use of the installed devbox-environment:
ENV \
	PATH="/code/.devbox/nix/profile/default/bin:${PATH}" \
	LD_LIBRARY_PATH="/code/.devbox/nix/profile/default/lib"

# 🚧 To-do: Check if “ENTRYPOINT” is not the better alternative, see:
# <https://www.bmc.com/blogs/docker-cmd-vs-entrypoint>
ENTRYPOINT ["devbox", "shell"]
# CMD ["devbox", "shell"]
