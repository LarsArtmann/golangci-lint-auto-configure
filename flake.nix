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

  outputs = { self, nixpkgs, flake-utils, goFindingSrc }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };

        version =
          if self ? rev
          then builtins.substring 0 7 self.rev
          else "dev";

        golangci-lint-auto-configure = pkgs.buildGoModule rec {
          pname = "golangci-lint-auto-configure";
          inherit version;

          src = ./.;

          vendorHash = pkgs.lib.fakeHash;

          subPackages = [ "cmd/golangci-lint-auto-configure" ];

          ldflags = [
            "-s"
            "-w"
            "-X main.version=${version}"
          ];

          env.CGO_ENABLED = 0;

          proxyVendor = true;

          postPatch = ''
            # Remove the local replace directive so go mod can resolve normally
            sed -i '/^replace github.com\/larsartmann\/go-finding/d' go.mod
            # Replace the dummy version with the actual commit from the flake input
            goFindingRev=$(cd ${goFindingSrc} && git rev-parse HEAD 2>/dev/null || echo "unknown")
            sed -i "s|github.com/larsartmann/go-finding v0.0.0-00010101000000-000000000000|github.com/larsartmann/go-finding v0.0.0-$(date -u -d @0 +%Y%m%d%H%M%S)-${goFindingRev}|" go.mod
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

        devShells.default = pkgs.mkShell {
          inputsFrom = [ golangci-lint-auto-configure ];

          packages = with pkgs; [
            go
            just
            golangci-lint
            ginkgo
            templ
            jq
            git
            pre-commit
            gopls
            gotools
            alejandra
          ];

          shellHook = ''
            echo "golangci-lint-auto-configure dev shell"
            echo "  Go:             $(go version)"
            echo "  golangci-lint:  $(golangci-lint version --short 2>/dev/null || echo 'N/A')"
            echo "  ginkgo:         $(ginkgo version 2>/dev/null || echo 'N/A')"
            echo "  templ:          $(templ version 2>/dev/null || echo 'N/A')"
            echo "  just:           $(just --version 2>/dev/null || echo 'N/A')"
          '';
        };

        checks = {
          build = golangci-lint-auto-configure;
        };
      }
    );
}
