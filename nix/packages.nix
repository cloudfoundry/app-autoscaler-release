{ buildGoModule
, buildGo123Module
, callPackage
, fetchFromGitHub
, fetchgit
, lib
}: {
  app-autoscaler-cli-plugin = buildGo123Module rec {
    pname = "app-autoscaler-cli-plugin";

    major = "4";
    minor = "0";
    patch = "1";
    version = "${major}.${minor}.${patch}";
    src = fetchgit {
      url = "https://github.com/cloudfoundry/app-autoscaler-cli-plugin";
      rev = "v${version}";
      hash = "sha256-si0tj8xPpkp7VbPDreEkqcudohWS7rlSPAr8GUtveUk=";
      fetchSubmodules = true;
    };
    doCheck = false;
    vendorHash = "sha256-/3bcnrn7KUUOXHKBe2WcnN9ozFQiNKXeNw2ctLKGUJg=";
    ldflags = ["-s" "-w" "-X 'main.BuildMajorVersion=${major}'" "-X 'main.BuildMinorVersion=${minor}'" "-X 'main.BuildPatchVersion=${patch}'"];
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

  cloud-mta-build-tool = buildGoModule rec {
    pname = "Cloud MTA Build Tool";
    version = "1.2.30";

    src = fetchFromGitHub {
      owner = "SAP";
      repo = "cloud-mta-build-tool";
      rev = "v${version}";
      hash = "sha256-iuNaaApnyfyqm3SvYG3en+a78MUP1BxSM3JZz+JhEFs=";
    };

    vendorHash = "sha256-pyXeuZGg3Yv6p8GNKC598EdZqX8KLc3rkewMkq4vA7c=";

    ldflags = ["-s" "-w" "-X main.Version=${version}"];

    doCheck = false;

    postInstall = ''
      pushd "''${out}/bin" &> /dev/null
        ln --symbolic 'cloud-mta-build-tool' 'mbt'
      popd
    '';
  };

  log-cache-cli-plugin = buildGoModule rec {
    pname = "log-cache-cli";
    version = "6.0.2";
    src = fetchgit {
      url = "https://github.com/cloudfoundry/log-cache-cli";
      rev = "v${version}";
      hash = "sha256-NhYpDxq5MhVOIMVulY1MG22cN3gaQi5agU7Aaw9Dr0A=";
      fetchSubmodules = true;
    };
    doCheck = false;
    vendorHash = null;
    ldflags = ["-s" "-w" "-X main.version=${version}"];
  };

  cf-deploy-plugin = buildGoModule rec {
    pname = "CF Deploy Plugin";
    version = "3.5.0";

    src = fetchFromGitHub {
      owner = "cloudfoundry";
      repo = "multiapps-cli-plugin";
      rev = "v${version}";
      hash = "sha256-SVPVPJJWOk08ivZWu9UwD9sIISajIukQpcFpc0tU1zg=";
    };
    vendorHash = "sha256-S066sNHhKxL4anH5qSSBngtOcAswopiYBXgKAvHyfAM=";
    CGO_ENABLED = 0;
    ldflags = [ "-w -X main.Version=${version}" ];
  };

  uaac = callPackage ./packages/uaac {};
}
