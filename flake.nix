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
  };

  outputs =
    inputs@{
      self,
      flake-parts,
      goFindingSrc,
      gogenfilterSrc,
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
          version = "0.2.0";
          commit = self.rev or "none";
          buildDate = self.lastModifiedDate or "unknown";
          ldflagsPkg = "github.com/larsartmann/golangci-lint-auto-configure/pkg/version";

          golangci-lint-auto-configure = pkgs.buildGoModule {
            pname = "golangci-lint-auto-configure";
            inherit version;

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

            proxyVendor = true;

            vendorHash = "sha256-1foCQuehiNwMhj4bdGTMxiJqL5I8YtYtarTDB/Z83Bs=";

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
            };

            # The replaces redirect private repos to SSH-fetched local sources.
            # go mod tidy needs network (fetches transitive deps of replaced
            # modules); only the go-modules FOD has network via __noChroot.
            # The sandboxed main derivation instead appends -mod=mod so Go
            # auto-reconciles go.mod from the FOD's proxy cache (no network).
            postPatch = ''
              echo 'replace github.com/larsartmann/go-finding => ${goFindingSrc}' >> go.mod
              echo 'replace github.com/LarsArtmann/gogenfilter/v3 => ${gogenfilterSrc}' >> go.mod
              if [[ "$name" == *go-modules* ]]; then
                export HOME="$TMPDIR"
                go mod tidy
              else
                # Append, don't overwrite — buildGoModule sets -trimpath
                # in GOFLAGS to prevent GOROOT leaking into the binary.
                export GOFLAGS+=" -mod=mod"
              fi
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
              };

              shellHook = ''
                export PATH="$HOME/go/bin:$PATH"
                echo "golangci-lint-auto-configure dev shell"
                echo "  Go:             $(go version)"
                echo "  golangci-lint:  $(golangci-lint version --short 2>/dev/null || echo 'N/A')"
                echo "  templ:          $(templ version 2>/dev/null || echo 'N/A')"
              '';
            };

            ci = pkgs.mkShellNoCC {
              packages = with pkgs; [
                go_1_26
                golangci-lint
                templ
              ];
              GOPRIVATE = "github.com/LarsArtmann";
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
