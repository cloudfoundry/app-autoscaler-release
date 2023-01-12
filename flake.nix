{
  description = "Dependencies of app-autoscaler-release";

  inputs = {
    nixpkgs.url = github:NixOS/nixpkgs/nixos-unstable;

    # The following input is needed until PR https://github.com/NixOS/nixpkgs/pull/189079
    # has been merged!
    # Alternative solution: produce a package here locally that contains cf-uaac
    # by making use of <https://nixos.org/manual/nixpkgs/stable/#developing-with-ruby>, see
    # chapter: 17.30.2.5. Packaging applications
    jdwpkgs.url = github:joergdw/nixpkgs/rubyPackages.cf-uaac;
  };

  outputs = { self, nixpkgs, jdwpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];

      # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      # Nixpkgs instantiated for supported system types.
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
      jdwpkgsFor = forAllSystems (system: import jdwpkgs { inherit system; });
    in {
      devShells = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
          jdwpkgs = jdwpkgsFor.${system};
        in {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [
              act
              bosh-cli
              cloudfoundry-cli
              fly
              ginkgo
              gh
              go
              golangci-lint
              google-cloud-sdk
              maven
              nodejs
              ruby
              jdwpkgs.rubyPackages.cf-uaac # comment for not compiling much
              shellcheck
              temurin-bin-11
            ];
          };
      });
  };
}