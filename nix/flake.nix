{
  description = "Extra dependencies of app-autoscaler-release for the devbox";

  inputs = {
    nixpkgs.url = github:NixOS/nixpkgs/nixos-unstable;
    nixpkgs-bosh-cli-v7-3-1.url = github:NixOS/nixpkgs/1179c6c3705509ba25bd35196fca507d2a227bd0;
  };

  outputs = { self, nixpkgs, nixpkgs-bosh-cli-v7-3-1 }:
    let
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];

      # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      # Nixpkgs instantiated for supported system types.
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });
      nixpkgsFor-bosh-cli-v7-3-1 = forAllSystems (system: import nixpkgs-bosh-cli-v7-3-1 { inherit system; });
    in {
      packages = forAllSystems (system:
        let
          nixpkgs = nixpkgsFor.${system};
          callPackages = nixpkgs.lib.customisation.callPackagesWith nixpkgs;
        in callPackages ./packages.nix {}
      );
  };
}