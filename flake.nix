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

        golangci-lint-auto-configure = pkgs.buildGoModule rec {
          pname = "golangci-lint-auto-configure";
          inherit version;

          src = pkgs.lib.cleanSourceWith {
            filter =
              path: _type:
              let
                b = baseNameOf path;
              in
              !(
                b == "vendor"
                || b == ".git"
                || b == "docs"
                || b == ".crush"
                || b == "reports"
                || b == "examples"
                || b == "scripts"
                || b == ".envrc"
                || b == ".github"
                || b == "bin"
                || b == "justfile"
                || b == "Dockerfile"
                || b == ".dockerignore"
                || b == ".gitattributes"
                || b == ".pre-commit-config.yaml"
                || b == ".pre-commit-hooks.yaml"
                || b == ".config"
                || pkgs.lib.hasSuffix ".md" b
                || pkgs.lib.hasSuffix ".lock" b
                || pkgs.lib.hasSuffix ".yml" b
                || pkgs.lib.hasSuffix ".yaml" b
              );
            src = pkgs.lib.cleanSource ./.;
          };

          proxyVendor = true;

          vendorHash = "sha256-eu2K277Cw6/0HDrFMHoXEesdjkNMRurJjbfT2gFedZM=";

          subPackages = [ "cmd/golangci-lint-auto-configure" ];

          ldflags = [
            "-s"
            "-w"
            "-X github.com/larsartmann/golangci-lint-auto-configure/pkg/version.version=${version}"
            "-X github.com/larsartmann/golangci-lint-auto-configure/pkg/version.commit=${commit}"
            "-X github.com/larsartmann/golangci-lint-auto-configure/pkg/version.date=${buildDate}"
            "-X github.com/larsartmann/golangci-lint-auto-configure/pkg/version.treeState=clean"
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
          program = "${golangci-lint-auto-configure}/bin/golangci-lint-auto-configure";
        };

        formatter = pkgs.nixfmt;

        devShells.default = pkgs.mkShell {
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
            # Install ginkgo from go.mod to ensure version match
            go install github.com/onsi/ginkgo/v2/ginkgo 2>/dev/null || true

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
        };
      }
    )
    // {
      overlays.default = _final: prev: {
        golangci-lint-auto-configure = self.packages.${prev.stdenv.system}.default;
      };
    };
}
