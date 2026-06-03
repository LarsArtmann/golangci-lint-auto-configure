{
  description = "Automatically configure and optimize golangci-lint configurations";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    goFindingSrc = {
      url = "git+ssh://git@github.com/LarsArtmann/go-finding?ref=master";
      flake = false;
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      goFindingSrc,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        version = "0.2.0";

        commit = self.rev or "none";

        buildDate = self.lastModifiedDate or "unknown";

        ldflagsPkg = "github.com/larsartmann/golangci-lint-auto-configure/pkg/version";

        golangci-lint-auto-configure = pkgs.buildGoModule {
          pname = "golangci-lint-auto-configure";
          inherit version;

          src = pkgs.lib.fileset.toSource {
            root = ./.;
            fileset = pkgs.lib.fileset.unions [
              ./go.mod
              ./go.sum
              ./cmd
              ./pkg
              ./internal
              ./scripts
            ];
          };

          proxyVendor = true;

          vendorHash = "sha256-LCz14+53dif4m6fq8I11hHkKwSueYKnjVjTl4EUQUl0=";

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

          postPatch = ''
            echo 'replace github.com/larsartmann/go-finding => ${goFindingSrc}' >> go.mod
          '';

          meta = with pkgs.lib; {
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
          program = pkgs.lib.getExe golangci-lint-auto-configure;
        };

        formatter = pkgs.nixfmt;

        devShells.default = pkgs.mkShellNoCC {
          packages = with pkgs; [
            go_1_26
            just
            golangci-lint
            templ
            jq
            git
            pre-commit
            gopls
            gotools
            nixfmt
          ];

          env = {
            CGO_ENABLED = "0";
            GOWORK = "off";
            GOPRIVATE = "github.com/LarsArtmann";
            GOTOOLCHAIN = "local";
          };

          shellHook = ''
            echo "golangci-lint-auto-configure dev shell"
            echo "  Go:             $(go version)"
            echo "  golangci-lint:  $(golangci-lint version --short 2>/dev/null || echo 'N/A')"
            echo "  ginkgo:         $(ginkgo version 2>/dev/null || echo 'N/A')"
            echo "  templ:          $(templ version 2>/dev/null || echo 'N/A')"
            echo "  just:           $(just --version)"
          '';
        };

        checks = {
          build = golangci-lint-auto-configure;
          test = golangci-lint-auto-configure.overrideAttrs (_: {
            doCheck = true;
          });
        };
      }
    )
    // {
      overlays.default = _final: prev: {
        golangci-lint-auto-configure = self.packages.${prev.stdenv.system}.default;
      };
    };
}
