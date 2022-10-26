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
              bosh-cli
              cloudfoundry-cli
              fly
              ginkgo
              go
              golangci-lint
              google-cloud-sdk
              maven
              nodejs
              ruby
              ## The following line needs: `nixpkgs.url = github:joergdw/nixpkgs/rubyPackages.cf-uaac;`
              ## until PR https://github.com/NixOS/nixpkgs/pull/189079 has been merged!
              # rubyPackages.cf-uaac
              shellcheck
              temurin-bin-11
            ];
          };
      });
  };
}