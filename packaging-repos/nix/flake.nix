{
  description = "Declarative Lightweight Static Site Generator";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = "0.1.0";
        srcUrl = {
          x86_64-linux = "garterscopic-linux-amd64";
          aarch64-linux = "garterscopic-linux-arm64";
          x86_64-darwin = "garterscopic-darwin-amd64";
          aarch64-darwin = "garterscopic-darwin-arm64";
        };
        binaryName = srcUrl.${system} or (throw "Unsupported system: ${system}");

        garterscopic = pkgs.stdenvNoCC.mkDerivation {
          pname = "garterscopic";
          inherit version;

          src = pkgs.fetchurl {
            url = "https://github.com/garterscopic/garterscopic/releases/download/v${version}/${binaryName}";
            sha256 = pkgs.lib.fakeSha256;
          };

          dontUnpack = true;

          installPhase = ''
            runHook preInstall
            install -Dm755 $src $out/bin/garterscopic
            runHook postInstall
          '';

          meta = with pkgs.lib; {
            description = "Declarative Lightweight Static Site Generator";
            longDescription = ''
              Garterscopic is a simple, fast static site generator that lets you
              build websites using HTML, YAML, Markdown, and CSS - no Node.js required.
            '';
            homepage = "https://garterscopic.dev";
            license = licenses.mit;
            maintainers = with maintainers; [ ];
            mainProgram = "garterscopic";
            platforms = [
              "x86_64-linux"
              "aarch64-linux"
              "x86_64-darwin"
              "aarch64-darwin"
            ];
          };
        };
      in
      {
        packages = {
          default = garterscopic;
          garterscopic = garterscopic;
        };

        apps = {
          default = {
            type = "app";
            program = "${self.packages.${system}.default}/bin/garterscopic";
          };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = [ garterscopic ];
        };
      }
    );
}
