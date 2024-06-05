{ buildGoModule
, callPackage
, fetchgit
, lib
}: {
  app-autoscaler-cli-plugin = buildGoModule rec {
    pname = "app-autoscaler-cli-plugin";
    gitCommit = "f46dc1ea62c4c7bd426c82f4e2a525b3a3c42300";
    version = "${gitCommit}";
    src = fetchgit {
      url = "https://github.com/cloudfoundry/app-autoscaler-cli-plugin";
      rev = "${gitCommit}";
      hash = "sha256-j8IAUhjYjEFvtRbA6o2vA7P2uUmKVYsd9uJmN0WtVCM=";
      fetchSubmodules = true;
    };
    doCheck = false;
    vendorHash = "sha256-NzEStcOv8ZQsHOA8abLABKy+ZE3/SiYbRD/ZVxo0CEk=";
  };

  # this bosh-bootloader custom build can be removed once https://github.com/cloudfoundry/bosh-bootloader/issues/596 is implemented.
  bosh-bootloader = buildGoModule rec {
    pname = "bosh-bootloader";
    version = "9.0.17";
    src = fetchgit {
      url = "https://github.com/cloudfoundry/bosh-bootloader";
      rev = "v${version}";
      fetchSubmodules = true;
      hash = "sha256-P4rS7Nv/09+9dD198z4NOXnldSE5fx3phEK24Acatps=";
    };
    doCheck = false;
    vendorHash = null;
  };

  log-cache-cli-plugin = buildGoModule rec {
    pname = "log-cache-cli";
    version = "6.0.1";
    src = fetchgit {
      url = "https://github.com/cloudfoundry/log-cache-cli";
      rev = "v${version}";
      hash = "sha256-XMxZPmqjOo/yaMFHY+zTjamB2FmPn2eh0zEtwQevt+I=";
      fetchSubmodules = true;
    };
    doCheck = false;
    vendorHash = null;
    ldflags = ["-s" "-w" "-X main.version=${version}"];
  };

  uaac = callPackage ./packages/uaac {};
}
