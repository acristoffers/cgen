{
  description = "Generate completion scripts from a CLI description";

  inputs =
    {
      nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

      flake-utils.url = "github:numtide/flake-utils";

      gitignore.url = "github:hercules-ci/gitignore.nix";
      gitignore.inputs.nixpkgs.follows = "nixpkgs";
    };

  outputs = inputs:
    let
      inherit (inputs) nixpkgs gitignore flake-utils;
      inherit (gitignore.lib) gitignoreSource;
    in
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      rec {
        formatter = pkgs.nixpkgs-fmt;
        packages.default = packages.cgen;
        packages.cgen = pkgs.buildGoModule {
          pname = "cgen";
          version = (builtins.readFile ./cgen/version);
          src = gitignoreSource ./.;
          vendorHash = "sha256-2mkxlI43ngiEu+5cd7gOw7bP44C6vbk3Q7RUYIiLub4=";
          buildInputs = with pkgs; [ glibc.static ];
          CFLAGS = "-I${pkgs.glibc.dev}/include";
          LDFLAGS = "-L${pkgs.glibc}/lib";
          ldflags = [ "-s" "-w" "-linkmode external" "-extldflags '-static'" ];
          installPhase = ''
            runHook preInstall
            mkdir -p $out/bin
            mkdir -p build
            $GOPATH/bin/docgen
            cp -r build/share $out/share
            cp $GOPATH/bin/cgen $out/bin/cgen
            runHook postInstall
          '';
        };
        apps = rec {
          cgen = { type = "app"; program = "${packages.cgen}/bin/cgen"; };
          default = cgen;
        };
        devShell = pkgs.mkShell {
          packages = with pkgs; [
            bat
            busybox
            git
            go
            man
          ];
        };
      }
    );
}
