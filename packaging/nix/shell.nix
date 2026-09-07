{ pkgs ? import <nixpkgs> { } }:

let
  garterscopic = pkgs.buildGoModule {
    pname = "garterscopic";
    version = "git";
    src = pkgs.lib.cleanSource ../..;

    vendorHash = "";

    CGO_ENABLED = 0;

    doCheck = false;

    meta = with pkgs.lib; {
      description = "Declarative Lightweight Static Site Generator";
      homepage = "https://github.com/italiatroller-1990/garterscopic";
      license = licenses.mit;
      maintainers = with maintainers; [ ];
      mainProgram = "garterscopic";
      platforms = platforms.linux ++ platforms.darwin ++ platforms.freebsd;
    };
  };
in
pkgs.mkShell {
  buildInputs = [ garterscopic ];

  shellHook = ''
    echo "Garterscopic development shell"
    echo "Version: $(garterscopic --version 2>/dev/null || echo 'built from source')"
  '';
}
