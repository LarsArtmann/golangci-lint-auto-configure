{
  description = "Automatically configure and optimize golangci-lint configurations";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    go-nix-helpers = {
      url = "git+https://github.com/LarsArtmann/go-nix-helpers?ref=master";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    go-finding = {
      url = "git+https://github.com/LarsArtmann/go-finding?ref=master";
      flake = false;
    };

    gogenfilter = {
      url = "github:LarsArtmann/gogenfilter?ref=refs/tags/v3.6.0";
      flake = false;
    };
  };

  outputs =
    inputs@{
      self,
      flake-parts,
      ...
    }:
    let
      version = self.rev or self.dirtyRev or "dev";
      commit = self.rev or "none";
      buildDate = self.lastModifiedDate or "unknown";
      ldflagsPkg = "github.com/larsartmann/golangci-lint-auto-configure/pkg/version";
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [ inputs.go-nix-helpers.flakeModules.go-standard ];

      go-standard = {
        pname = "golangci-lint-auto-configure";
        vendorHash = import ./vendorHash.nix;
        description = "Automatically configure and optimize golangci-lint configurations";
        enableTempl = true;

        deps = {
          "github.com/larsartmann/go-finding" = inputs.go-finding;
          "github.com/LarsArtmann/gogenfilter/v3" = inputs.gogenfilter;
        };
        validatePrivateDeps = false;

        src = inputs.nixpkgs.lib.fileset.toSource {
          root = ./.;
          fileset = inputs.nixpkgs.lib.fileset.unions [
            ./go.mod
            ./go.sum
            ./cmd
            ./pkg
            ./internal
            ./scripts
          ];
        };

        subPackages = [
          "cmd/golangci-lint-auto-configure"
          "cmd/coverage-check"
        ];

        ldflags = [
          "-s"
          "-w"
          "-X ${ldflagsPkg}.version=${version}"
          "-X ${ldflagsPkg}.commit=${commit}"
          "-X ${ldflagsPkg}.date=${buildDate}"
          "-X ${ldflagsPkg}.treeState=clean"
        ];

        extraBuildAttrs = {
          env = {
            CGO_ENABLED = "0";
            GOEXPERIMENT = "jsonv2";
          };
        };

        shellExtraEnv = {
          GOPRIVATE = "github.com/larsartmann/*,github.com/LarsArtmann/*";
          GOEXPERIMENT = "jsonv2";
          CGO_ENABLED = "0";
        };

        devShellExtraPackages = pkgs: [
          pkgs.jq
          pkgs.git
          pkgs.pre-commit
          pkgs.gcc
          pkgs.markdownlint-cli2
        ];

        enableNixfmt = true;
      };

      perSystem =
        { config, pkgs, ... }:
        {
          checks.race = config.packages.default.overrideAttrs (old: {
            doCheck = true;
            checkFlags = [ "-race" ];
            env = old.env // {
              CGO_ENABLED = "1";
            };
          });

          apps.coverage-check = {
            type = "app";
            program = "${config.packages.default}/bin/coverage-check";
            meta.description = "Enforce minimum test coverage threshold from a Go coverage profile";
          };

          apps.lint =
            let
              lintScript = pkgs.writeShellScript "repo-lint" ''
                set -euo pipefail
                export CGO_ENABLED=1
                export GOEXPERIMENT=jsonv2
                export GOPRIVATE=github.com/larsartmann/*,github.com/LarsArtmann/*
                export GOCACHE="''${XDG_CACHE_HOME:-$HOME/.cache}/golangci-lint-auto-configure-lint"
                cd "$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
                exec ${pkgs.golangci-lint}/bin/golangci-lint run --config=.golangci.yml --timeout=10m "$@"
              '';
            in
            {
              type = "app";
              program = "${lintScript}";
              meta.description = "Run golangci-lint over the repo with the committed config";
            };
        };
    };
}
