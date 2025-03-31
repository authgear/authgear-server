{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in
      {
        devShells.default = pkgs.mkShell {
          packages = [
            # 1.24.1
            pkgs.go
            # Any nodejs 20 is fine.
            pkgs.nodejs_20
            # The version of python3 does not matter that much.
            pkgs.python3

            pkgs.pkg-config

            (pkgs.golangci-lint.overrideAttrs (
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

            (pkgs.buildGoModule {
              name = "wire";
              src = pkgs.fetchFromGitHub {
                owner = "google";
                repo = "wire";
                rev = "v0.5.0";
                sha256 = "sha256-9xjymiyPFMKbysgZULmcBEMI26naUrLMgTA+d7Q+DA0=";
              };
              vendorHash = "sha256-ZFUX4LgPte6oAf94D82Man/P9VMpx+CDNCTMBwiy9Fc=";
              subPackages = [ "cmd/wire" ];
            })

            (pkgs.buildGoModule {
              name = "govulncheck";
              src = pkgs.fetchgit {
                url = "https://go.googlesource.com/vuln";
                rev = "refs/tags/v1.1.4";
                hash = "sha256-d1JWh/K+65p0TP5vAQbSyoatjN4L5nm3VEA+qBSrkAA=";
              };
              vendorHash = "sha256-MSTKDeWVxD2Fa6fNoku4EwFwC90XZ5acnM67crcgXDg=";
              subPackages = [ "cmd/govulncheck" ];
              # checkPhase by default run tests. Running tests will result in build error.
              # So we skip it.
              doCheck = false;
            })

            (pkgs.buildGoModule {
              name = "goimports";
              src = pkgs.fetchgit {
                url = "https://go.googlesource.com/tools";
                rev = "refs/tags/v0.29.0";
                hash = "sha256-h3UjRY1w0AyONADNiLhxXt9/z7Tb/40FJI8rKGXpBeM=";
              };
              vendorHash = "sha256-5JE4/2yn5emRuq4wGTeNnAJa+QZayDq2n+dL3gMML/M=";
              subPackages = [ "cmd/goimports" ];
            })

            (pkgs.buildGoModule {
              name = "xk6";
              src = pkgs.fetchFromGitHub {
                owner = "grafana";
                repo = "xk6";
                rev = "v0.14.0";
                sha256 = "sha256-s68kM8HPl+hSsiXbO1nO3D3R8MN9fABc9M+30jDCfek=";
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
