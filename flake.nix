{
  description = "Awtrix Listens To Altar Requests";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = {
    self,
    nixpkgs,
  }: let
    systems = [
      "x86_64-linux"
      "aarch64-linux"
      "x86_64-darwin"
      "aarch64-darwin"
    ];
    forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
  in {
    packages = forAllSystems (pkgs: {
      default = pkgs.buildGoModule {
        pname = "altar";
        version = "1.1.2";
        src = self;

        # No external dependencies (stdlib only) => no vendoring needed.
        vendorHash = null;

        # Embed the IANA tz database so time.LoadLocation works without a
        # runtime dependency on system zoneinfo (important under the hardened
        # systemd sandbox).
        tags = ["timetzdata"];

        meta = {
          description = "Awtrix Listens To Altar Requests";
          mainProgram = "altar";
        };
      };
    });

    nixosModules.default = import ./nix/module.nix self;

    devShells = forAllSystems (pkgs: {
      default = pkgs.mkShell {
        packages = [pkgs.go];
      };
    });
  };
}
