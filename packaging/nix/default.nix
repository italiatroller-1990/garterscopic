{ lib
, buildGoModule
, fetchFromGitHub
}:

buildGoModule rec {
  pname = "garterscopic";
  version = "{{VERSION}}";

  src = fetchFromGitHub {
    owner = "italiatroller-1990";
    repo = "garterscopic";
    rev = "v${version}";
    hash = "";  # To be filled by nix-build or nix-update
  };

  vendorHash = null;

  CGO_ENABLED = 0;

  ldflags = [
    "-s"
    "-w"
    "-X" "main.version=${version}"
  ];

  doCheck = false;

  meta = with lib; {
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
}
