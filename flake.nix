{
  description = "Dependencies of app-autoscaler-release";

  inputs = {
    nixpkgs.url = github:NixOS/nixpkgs/nixos-23.05;
  };

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];

      # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      # Nixpkgs instantiated for supported system types.
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
    in {
      packages = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in {
          swagger-cli = pkgs.buildNpmPackage rec {
            pname = "swagger-cli";
            version = "4.0.5";

            src = pkgs.fetchFromGitHub {
              owner = "empire-medical";
              repo = pname;
              rev = "d2a9e4d9b6675a6003ba74669e69df23db979e07";
              hash = "sha256-fvKWQibOume+r3ScLTxJMapdD/FUtKh9V6gBHH5dL7o="; # This is already correct!
            };
            # npmDeps = pkgs.fetchNpmDeps {
            #   inherit forceGitDeps src srcs sourceRoot prePatch patches postPatch;
            #   name = "${name}-npm-deps";
            #   hash = npmDepsHash;
            # };

            npmDepsHash = "sha256-go9eYGCZmbwRArHVTVa6mxL+kjvBcrLxKw2iVv0a5hY=";

            # # The prepack script runs the build script, which we'd rather do in the build phase.
            npmFlags = [ "--legacy-peer-deps" ];
            makeCacheWritable = true;

            # # NODE_OPTIONS = "--openssl-legacy-provider";
            # src = pkgs.fetchFromGitHub {
            #   owner = "APIDevTools";
            #   repo = "swagger-cli";
            #   rev = "v${version}";
            #   sha256 = "sha256-WgzfSd57vRwa1HrSgNxD0F5ckczBkOaVmrEZ9tMAcRA=";
            # };

            # npmDepsHash = "sha256-go9eYGCZmbwRArHVTVa6mxL+kjvBcrLxKw2iVv0a5hY=";

            buildPhase = ''
              npm run bump
            '';

            meta = {
              description = ''
                Validate Swagger/OpenAPI files in JSON or YAML format
                Supports multi-file API definitions via $ref pointers
                Bundle multiple Swagger/OpenAPI files into one combined file
              '';
              homepage = "<https://github.com/empire-medical/swagger-cli>";
              # license = licenses.mit;
            };
          };
      });

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
              delve # go-debugger
              fly
              ginkgo
              gh
              gnumake
              go
              go-tools
              golangci-lint
              gopls # See: <https://github.com/golang/vscode-go/blob/master/docs/tools.md>
              google-cloud-sdk
              maven
              nodejs
              # # We use the binary-buildpack and nix-build-results have set the wrong ELF-Interpreter.
              # # For more background, see:
              # # <https://blog.thalheim.io/2022/12/31/nix-ld-a-clean-solution-for-issues-with-pre-compiled-executables-on-nixos>
              # patchelf
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
              # swagger-cli
              temurin-bin
              yq-go
            ];

            # Needed to run with delve, see: <https://nixos.wiki/wiki/Go#Using_cgo_on_NixOS>
            hardeningDisable = [ "fortify" ];

            shellHook = ''
              echo -ne '\033[1;33m' '\033[5m'
              cat << 'EOF'
                ⚠️ If `whoami` does not work properly on your computer, `bosh ssh` commands may fail.
                The solution is, to provide your nix dev-shell the path to the `libnss_sss.so.2` of
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