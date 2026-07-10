{
  description = "Automatically configure and optimize golangci-lint configurations";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };

    systems.url = "github:nix-systems/default";

    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    # Private LarsArtmann repos fetched via SSH. The Go module proxy does not
    # cache these, so we inject replace directives pointing to these sources.
    goFindingSrc = {
      url = "git+ssh://git@github.com/LarsArtmann/go-finding?ref=master";
      flake = false;
    };
    gogenfilterSrc = {
      url = "git+ssh://git@github.com/LarsArtmann/gogenfilter?ref=master";
      flake = false;
    };
    go-nix-helpers = {
      url = "git+ssh://git@github.com/LarsArtmann/go-nix-helpers?ref=master";
      flake = false;
    };
  };

  outputs =
    inputs@{
      self,
      flake-parts,
      goFindingSrc,
      gogenfilterSrc,
      go-nix-helpers,
      ...
    }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import inputs.systems;

      imports = [ inputs.treefmt-nix.flakeModule ];

      perSystem =
        {
          config,
          pkgs,
          lib,
          ...
        }:
        let
          version = self.rev or self.dirtyRev or "dev";
          commit = self.rev or "none";
          buildDate = self.lastModifiedDate or "unknown";
          ldflagsPkg = "github.com/larsartmann/golangci-lint-auto-configure/pkg/version";

          mkPreparedSource = import (go-nix-helpers + "/mkPreparedSource.nix") {
            inherit pkgs lib;
            goPkg = pkgs.go_1_26;
          };

          src = lib.fileset.toSource {
            root = ./.;
            fileset = lib.fileset.unions [
              ./go.mod
              ./go.sum
              ./cmd
              ./pkg
              ./internal
              ./scripts
            ];
          };

          preparedSrc = mkPreparedSource {
            name = "golangci-lint-auto-configure";
            inherit version src;
            deps = {
              "github.com/larsartmann/go-finding" = goFindingSrc;
              "github.com/LarsArtmann/gogenfilter/v3" = gogenfilterSrc;
            };
            validatePrivateDeps = false;
          };

          golangci-lint-auto-configure = pkgs.buildGoModule {
            pname = "golangci-lint-auto-configure";
            inherit version;

            src = preparedSrc;

            proxyVendor = true;

            vendorHash = "sha256-XLckeSqONjQS0TQGuVTCkyDhrxKXx3gIRlgpGIbtXqw=";

            subPackages = [ "cmd/golangci-lint-auto-configure" ];

            ldflags = [
              "-s"
              "-w"
              "-X ${ldflagsPkg}.version=${version}"
              "-X ${ldflagsPkg}.commit=${commit}"
              "-X ${ldflagsPkg}.date=${buildDate}"
              "-X ${ldflagsPkg}.treeState=clean"
            ];

            env = {
              CGO_ENABLED = 0;
              GOWORK = "off";
              GOEXPERIMENT = "jsonv2";
            };

            overrideModAttrs = _: {
              preBuild = ''
                export HOME=$TMPDIR
                go mod tidy
              '';
            };

            preBuild = ''
              export GOFLAGS+=" -mod=mod"
            '';

            meta = with lib; {
              description = "Automatically configure and optimize golangci-lint configurations";
              homepage = "https://github.com/LarsArtmann/golangci-lint-auto-configure";
              license = licenses.mit;
              mainProgram = "golangci-lint-auto-configure";
              platforms = platforms.unix;
            };
          };
        in
        {
          packages.default = golangci-lint-auto-configure;

          apps.default = {
            type = "app";
            program = lib.getExe golangci-lint-auto-configure;
          };

          devShells = {
            default = pkgs.mkShellNoCC {
              packages = with pkgs; [
                go_1_26
                golangci-lint
                templ
                jq
                git
                pre-commit
                gopls
                gotools
                nixfmt
                gcc
              ];

              env = {
                CGO_ENABLED = "0";
                GOWORK = "off";
                GOPRIVATE = "github.com/LarsArtmann";
                GOTOOLCHAIN = "local";
                GOEXPERIMENT = "jsonv2";
              };

              shellHook = ''
                export PATH="$HOME/go/bin:$PATH"
                echo "golangci-lint-auto-configure dev shell"
                echo "  Go:             $(go version)"
                echo "  golangci-lint:  $(golangci-lint version --short 2>/dev/null || echo 'N/A')"
                echo "  templ:          $(templ version 2>/dev/null || echo 'N/A')"
                echo "  GOEXPERIMENT:   $GOEXPERIMENT"
              '';
            };

            ci = pkgs.mkShellNoCC {
              packages = with pkgs; [
                go_1_26
                golangci-lint
                templ
              ];
              GOPRIVATE = "github.com/LarsArtmann";
              GOEXPERIMENT = "jsonv2";
            };
          };

          checks = {
            format = config.treefmt.build.check self;
            build = golangci-lint-auto-configure;
            test = golangci-lint-auto-configure.overrideAttrs (_: {
              doCheck = true;
            });
            race = golangci-lint-auto-configure.overrideAttrs (old: {
              doCheck = true;
              checkFlags = [ "-race" ];
              env = old.env // {
                CGO_ENABLED = "1";
              };
            });
          };

          treefmt = {
            projectRootFile = "go.mod";
            programs = {
              nixfmt.enable = true;
              gofumpt.enable = true;
              goimports.enable = true;
              templ.enable = true;
            };
          };
        };

      flake.overlays.default = final: _prev: {
        golangci-lint-auto-configure = self.packages.${final.stdenv.system}.default;
      };
    };
}
