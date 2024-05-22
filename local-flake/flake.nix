{
  description = "Dependencies of app-autoscaler-release";

  inputs = {
    nixpkgs.url = github:NixOS/nixpkgs/nixos-unstable;
  };

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];

      # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      # Nixpkgs instantiated for supported system types.
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in {
      packages = forAllSystems (system: {
        app-autoscaler-cli-plugin = nixpkgsFor.${system}.buildGoModule rec {
          pname = "app-autoscaler-cli-plugin";
          gitCommit = "f46dc1ea62c4c7bd426c82f4e2a525b3a3c42300";
          version = "${gitCommit}";
          src = nixpkgsFor.${system}.fetchgit {
            url = "https://github.com/cloudfoundry/app-autoscaler-cli-plugin";
            rev = "${gitCommit}";
            hash = "sha256-j8IAUhjYjEFvtRbA6o2vA7P2uUmKVYsd9uJmN0WtVCM=";
            fetchSubmodules = true;
          };
          doCheck = false;
          vendorHash = "sha256-NzEStcOv8ZQsHOA8abLABKy+ZE3/SiYbRD/ZVxo0CEk=";
        };

        # this bosh-bootloader custom build can be removed once https://github.com/cloudfoundry/bosh-bootloader/issues/596 is implemented.
        bosh-bootloader = nixpkgsFor.${system}.buildGoModule rec {
          pname = "bosh-bootloader";
          version = "9.0.17";
          src = nixpkgsFor.${system}.fetchgit {
            url = "https://github.com/cloudfoundry/bosh-bootloader";
            rev = "v${version}";
            fetchSubmodules = true;
            hash = "sha256-P4rS7Nv/09+9dD198z4NOXnldSE5fx3phEK24Acatps=";
          };
          doCheck = false;
          vendorHash = null;
        };

        log-cache-cli-plugin = nixpkgsFor.${system}.buildGoModule rec {
          pname = "log-cache-cli";
          version = "6.0.1";
          src = nixpkgsFor.${system}.fetchgit {
            url = "https://github.com/cloudfoundry/log-cache-cli";
            rev = "v${version}";
            hash = "sha256-XMxZPmqjOo/yaMFHY+zTjamB2FmPn2eh0zEtwQevt+I=";
            fetchSubmodules = true;
          };
          doCheck = false;
          vendorHash = null;
          ldflags = ["-s" "-w" "-X main.version=${version}"];
        };

        mbt = nixpkgsFor.${system}.buildGoModule rec {
          pname = "Cloud MTA Build Tool";
          version = "1.2.26";

          src = nixpkgsFor.${system}.fetchFromGitHub {
            owner = "SAP";
            repo = "cloud-mta-build-tool";
            rev = "v${version}";
            hash = "sha256-DKZ9Nj/sNC9dRjyiu4MKjLrIJWluYlZzUHWqEqtrNt4=";
          };

          vendorHash = "sha256-h8LPsuxvbr/aRhH1vR1fYgBot37yrfiemZTJMKj0zbk=";

          ldflags = ["-s" "-w" "-X main.Version=${version}"];

          doCheck = false;

          postInstall = ''
            pushd "$out/bin" &> /dev/null
              ln -s 'cloud-mta-build-tool' 'mbt'
            popd
          '';
        };

        sap-piper = nixpkgsFor.${system}.buildGoModule rec {
          pname = "Project Piper";
          version = "1.360.0";

          src = nixpkgsFor.${system}.fetchFromGitHub {
            owner = "SAP";
            repo = "jenkins-library";
            rev = "v${version}";
            hash = "sha256-eCg9fgP4pIhbW8zkg6izHb6V1VrkA+4hUCNvjMYOagI=";
          };

          vendorHash = "sha256-qCoz6jGoqGYsTYumrdy4RA3LdVwaMhS+vcCK2n970eg=";

          ldflags = ["-s" "-w" "-X github.com/SAP/jenkins-library/cmd.GitTag=v${version}"];

          doCheck = false;

          postInstall = ''
            pushd "$out/bin" &> /dev/null
              ln -s 'jenkins-library' 'piper'
            popd
          '';
        };

        uaac =  nixpkgsFor.${system}.bundlerApp rec {
          pname = "cf-uaac";
          gemdir = ./.;
          exes = ["uaac"];

          meta = {
            description = "CloudFoundry UAA Command Line Client";
            homepage = "https://github.com/cloudfoundry/cf-uaac";
            mainProgram = "uaac";
          };
        };
      });
  };
}
