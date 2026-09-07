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
    hash = "";
  };

  vendorHash = "";

  CGO_ENABLED = 0;

  doCheck = false;

  meta = with lib; {
    description = "Declarative Lightweight Static Site Generator";
    longDescription = ''
      Garterscopic is a simple, fast static site generator that lets you
      build websites using HTML, YAML, Markdown, and CSS - no Node.js required.
    '';
    homepage = "https://github.com/italiatroller-1990/garterscopic";
    license = licenses.mit;
    maintainers = with maintainers; [ garterscopic ];
    mainProgram = "garterscopic";
    platforms = platforms.linux ++ platforms.darwin ++ platforms.freebsd;
  };
}
