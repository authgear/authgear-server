{
  description = "A basic flake with a shell";
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    systems.url = "github:nix-systems/default";
    flake-utils = {
      url = "github:numtide/flake-utils";
      inputs.systems.follows = "systems";
    };
    go_1_23_4.url = "github:NixOS/nixpkgs/de1864217bfa9b5845f465e771e0ecb48b30e02d";
    nodejs_20_9_0.url = "github:NixOS/nixpkgs/a71323f68d4377d12c04a5410e214495ec598d4c";
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      go_1_23_4,
      nodejs_20_9_0,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        go_pkgs = go_1_23_4.legacyPackages.${system};
        go = go_pkgs.go;
        nodejs = nodejs_20_9_0.legacyPackages.${system}.nodejs_20;
      in
      {
        devShells.default = pkgs.mkShell {
          packages = [
            go
            nodejs
            # The version of python3 does not matter that much.
            pkgs.python3

            pkgs.pkg-config

            (go_pkgs.golangci-lint.overrideAttrs (
              prev:
              let
                version = "1.63.4";
              in
              {
                inherit version;
                src = pkgs.fetchFromGitHub {
                  owner = "golangci";
                  repo = "golangci-lint";
                  rev = "v${version}";
                  hash = "sha256-7nIo6Nuz8KLuQlT7btjnTRFpOl+KVd30v973HRKzh08=";
                };
                vendorHash = "sha256-atr4HMxoPEfGeaNlHqwTEAcvgbSyzgCe262VUg3J86c=";
                # We do not actually override anything here,
                # but if we do not repeat this, ldflags refers to the original version.
                ldflags = [
                  "-s"
                  "-X main.version=${version}"
                  "-X main.commit=v${version}"
                  "-X main.date=19700101-00:00:00"
                ];
              }
            ))

            (go_pkgs.buildGoModule {
              name = "mockgen";
              src = go_pkgs.fetchFromGitHub {
                owner = "golang";
                repo = "mock";
                rev = "v1.6.0";
                sha256 = "sha256-5Kp7oTmd8kqUN+rzm9cLqp9nb3jZdQyltGGQDiRSWcE=";
              };
              subPackages = [ "mockgen" ];
              vendorHash = "sha256-5gkrn+OxbNN8J1lbgbxM8jACtKA7t07sbfJ7gVJWpJM=";
            })

            (go_pkgs.buildGoModule {
              name = "wire";
              src = go_pkgs.fetchFromGitHub {
                owner = "google";
                repo = "wire";
                rev = "v0.5.0";
                sha256 = "sha256-9xjymiyPFMKbysgZULmcBEMI26naUrLMgTA+d7Q+DA0=";
              };
              vendorHash = "sha256-ZFUX4LgPte6oAf94D82Man/P9VMpx+CDNCTMBwiy9Fc=";
              subPackages = [ "cmd/wire" ];
            })

            (go_pkgs.buildGoModule {
              name = "govulncheck";
              src = go_pkgs.fetchgit {
                url = "https://go.googlesource.com/vuln";
                rev = "refs/tags/v1.1.3";
                hash = "sha256-ydJ8AeoCnLls6dXxjI05+THEqPPdJqtAsKTriTIK9Uc=";
              };
              vendorHash = "sha256-jESQV4Na4Hooxxd0RL96GHkA7Exddco5izjnhfH6xTg=";
              subPackages = [ "cmd/govulncheck" ];
              # checkPhase by default run tests. Running tests will result in build error.
              # So we skip it.
              doCheck = false;
            })

            (go_pkgs.buildGoModule {
              name = "goimports";
              src = go_pkgs.fetchgit {
                url = "https://go.googlesource.com/tools";
                rev = "refs/tags/v0.28.0";
                hash = "sha256-BCxsVz4f2h75sj1LzDoKvQ9c8P8SYjcaQE9CdzFdt3w=";
              };
              vendorHash = "sha256-MSir25OEmQ7hg0OAOjZF9J5a5SjlJXdOc523uEBSOSs=";
              subPackages = [ "cmd/goimports" ];
            })

            (go_pkgs.buildGoModule {
              name = "xk6";
              src = go_pkgs.fetchFromGitHub {
                owner = "grafana";
                repo = "xk6";
                rev = "v0.13.3";
                sha256 = "sha256-lmtGljTLbcOkE+CYupocM9gmHsTVnpPT9sXOKVuFOww=";
              };
              vendorHash = null;
              doCheck = false;
              subPackages = [ "cmd/xk6" ];
            })
          ];
          buildInputs = [
            pkgs.icu
            pkgs.vips
            # file includes libmagic.
            pkgs.file
          ];
          nativeBuildInputs = [
            pkgs.pkg-config
            pkgs.icu
            pkgs.vips
            pkgs.file
          ];
        };
      }
    );
}
