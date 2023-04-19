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
      devShells = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              act
              actionlint
              bosh-cli
              cloudfoundry-cli
              fly
              ginkgo
              gh
              gnumake
              go
              golangci-lint
              gopls # See: <https://github.com/golang/vscode-go/blob/master/docs/tools.md>
              google-cloud-sdk
              maven
              nodejs
              ruby
              # The following package for cf-uaac is needed by our makefile as well.
              # Until PR https://github.com/NixOS/nixpkgs/pull/189079 has been merged, this requires
              # as additional input: `jdwpkgs.url = github:joergdw/nixpkgs/rubyPackages.cf-uaac;`
              # Alternative solution 1: `gem install …` using https://direnv.net/man/direnv-stdlib.1.html#codelayout-rubycode
              # to create a project-specific ruby-gem-path – This solution is currently applied!
              #
              # Alternative solution 2: produce a package here locally that contains cf-uaac
              # by making use of <https://nixos.org/manual/nixpkgs/stable/#developing-with-ruby>, see
              # chapter: 17.30.2.5. Packaging applications
              #
              # jdwpkgs.rubyPackages.cf-uaac
              shellcheck
              temurin-bin
              yq-go
            ];
          };
      });
  };
}