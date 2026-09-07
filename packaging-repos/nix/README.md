# Garterscopic Nix Flake

This repository contains the [Nix](https://nixos.org/) flake for [Garterscopic](https://github.com/garterscopic/garterscopic).

## Usage

### NixOS Configuration

```nix
{
  inputs.garterscopic.url = "github:garterscopic/garterscopic-nix";

  outputs = { self, nixpkgs, garterscopic }: {
    nixosConfigurations.myHost = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        ({ pkgs, ... }: {
          environment.systemPackages = [ garterscopic.packages.${pkgs.system}.garterscopic ];
        })
      ];
    };
  };
}
```

### Run Without Installing

```bash
nix run github:garterscopic/garterscopic-nix -- --help
```

### Home Manager

```nix
{
  inputs.garterscopic.url = "github:garterscopic/garterscopic-nix";

  outputs = { self, nixpkgs, home-manager, garterscopic, ... }: {
    homeConfigurations."user@host" = home-manager.lib.homeManagerConfiguration {
      pkgs = nixpkgs.legacyPackages.x86_64-linux;
      modules = [
        ({ pkgs, ... }: {
          home.packages = [ garterscopic.packages.${pkgs.system}.garterscopic ];
        })
      ];
    };
  };
}
```

## Structure

```
.
├── .github/workflows/
│   └── check.yml        # Validates the flake
├── flake.nix            # Main flake definition
├── flake.lock           # Locked dependencies
├── README.md
└── default.nix          # Package definition
```

## How It Works

The flake automatically follows the latest release of Garterscopic. It supports:

- Linux (x86_64, aarch64)
- macOS (x86_64, aarch64)
- FreeBSD (x86_64, aarch64)
