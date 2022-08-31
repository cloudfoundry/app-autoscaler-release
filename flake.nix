{
  description = "Dependencies of app-autoscaler-release";

  inputs = {
    # Adapt later, after PR https://github.com/joergdw/nixpkgs/compare/rubyPackages.cf-uaac?expand=1 has been merged!
    nixpkgs.url = github:joergdw/nixpkgs/rubyPackages.cf-uaac;
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
              rubyPackages.cf-uaac
            ];
          };
      });
  };
}