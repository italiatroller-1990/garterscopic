{
  description = "Nix flake for Garterscopic - Declarative Lightweight Static Site Generator";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = builtins.replaceStrings [ "v" ] [ "" ] (builtins.head (builtins.split "\n" (builtins.readFile ./go.mod)));
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "garterscopic";
          inherit version;

          src = pkgs.lib.cleanSource ./.;

          vendorHash = null;

          CGO_ENABLED = 0;

          ldflags = [
            "-s"
            "-w"
            "-X" "main.version=${version}"
          ];

          meta = with pkgs.lib; {
            description = "Declarative Lightweight Static Site Generator";
            longDescription = ''
              Garterscopic is a simple, fast static site generator that lets you
              build websites using HTML, YAML, Markdown, and CSS - no Node.js required.
            '';
            homepage = "https://github.com/italiatroller-1990/garterscopic";
            license = licenses.asl20;
            maintainers = with maintainers; [ ];
            mainProgram = "garterscopic";
            platforms = platforms.linux ++ platforms.darwin ++ platforms.freebsd;
          };
        };

        packages.src-dist = let
          version = builtins.replaceStrings [ "v" ] [ "" ] (builtins.head (builtins.split "\n" (builtins.readFile ./go.mod)));
        in pkgs.stdenv.mkDerivation {
          pname = "garterscopic-src";
          inherit version;
          src = pkgs.lib.cleanSource ./.;
          phases = [ "installPhase" ];
          installPhase = ''
            mkdir -p $out
            cp -r $src/* $out/
          '';
        };

        apps.default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/garterscopic";
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [ go ];
        };

        checks = {
          build = self.packages.${system}.default;
        };
      });
}
