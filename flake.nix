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

          src = builtins.path {
            path = ./.;
            name = "source";
            filter = path: type:
              let
                baseName = baseNameOf path;
              in
              !(
                builtins.match "^(result|result-.*|bin|coverage|coverage\\.html|\\.idea|\\.vscode|node_modules)$" baseName != null
              );
          };

          vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";

          subPackages = [ "cmd/golangci-lint-auto-configure" ];

          ldflags = [
            "-s"
            "-w"
            "-X main.version=${version}"
          ];

          env.CGO_ENABLED = 0;

          GOWORK = "off";

          modRoot = ".";

          postPatch = ''
            # Remove the local replace directive
            sed -i '/^replace github.com\/larsartmann\/go-finding/d' go.mod
            # Rewrite the go-finding requirement to use a pseudo-version
            sed -i 's/github.com\/larsartmann\/go-finding v0.0.0-00010101000000-000000000000/github.com\/larsartmann\/go-finding v0.0.0-00010101000000-000000000000/' go.mod
          '';

          postConfigure = ''
            # Inject go-finding from flake input into vendor directory
            rm -rf vendor/github.com/larsartmann/go-finding
            cp -r ${goFindingSrc} vendor/github.com/larsartmann/go-finding
            chmod -R u+w vendor/github.com/larsartmann/go-finding
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
