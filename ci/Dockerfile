#  Debian-based!
FROM jetpackio/devbox:latest

# Installing your devbox project
WORKDIR /code
USER root:root
RUN <<-EODI
	mkdir --parents '/code'
	chown "${DEVBOX_USER}:${DEVBOX_USER}" '/code'
EODI
USER ${DEVBOX_USER}:${DEVBOX_USER}

COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} devbox.json devbox.json
COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} devbox.lock devbox.lock
COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} nix nix

# 💡 In theory we could avoid the copies by mounting the content into the image during
# build-time. However devbox has issues with the read-only-mounts in the name of the user
# “root”.
RUN \
	# --mount=type=bind,source=./nix,target=/code/nix,readonly \
	# --mount=type=bind,source=./devbox.json,target=/code/debox.json,readonly \
	# --mount=type=bind,source=./devbox.lock,target=/code/devbox.lock,readonly \
	devbox run -- echo "Installed Packages."

# Make use of the installed devbox-environment: This is required for the defined entry-point to
# work.
ENV \
	PATH="/code/.devbox/nix/profile/default/bin:${PATH}" \
	LD_LIBRARY_PATH="/code/.devbox/nix/profile/default/lib"

# The image is meant to be run with the particular devbox-setup from the “code”-folder.
ENTRYPOINT ["devbox", "run", "--config=/code", "--"]

# Avoid the bash from debox.json and use Debian's bash for better pre-configuration.
CMD ["/usr/bin/bash"]
