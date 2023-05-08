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
              # We use the binary-buildpack and nix-build-results have set the wrong ELF-Interpreter.
              # For more background, see:
              # <https://blog.thalheim.io/2022/12/31/nix-ld-a-clean-solution-for-issues-with-pre-compiled-executables-on-nixos>
              patchelf
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

            shellHook = ''
              echo -ne '\033[0;33m'
              cat << 'EOF'
                If `whoami` does not work properly on your computer, `bosh ssh` commands may fail.
                The solution is to provide your nix dev-shell the path to the `libnss_sss.so.2` of
                your host system, see: <https://github.com/NixOS/nixpkgs/issues/230110>

                Adapt the following line to contain the correct path:
                export LD_PRELOAD='/lib/x86_64-linux-gnu/libnss_sss.so.2'
              EOF
              echo -ne '\033[0m'
            '';
          };
      });
  };
}