let pkgs = import <nixpkgs> {};
in

{ stdenv         ? pkgs.stdenv
, buildGoPackage ? pkgs.buildGoPackage
, fetchgit       ? pkgs.fetchgit
, libsodium      ? pkgs.libsodium
, zeromq         ? pkgs.zeromq
, glide          ? pkgs.glide
, git            ? pkgs.git
, pkgconfig      ? pkgs.pkgconfig
, withZMQ        ? false }:

buildGoPackage rec {
  name = "skygear";

  goPackagePath = "github.com/skygeario/skygear-server";

  # If you are doing development and want to build the current source:
  src = ./.;

  # If you are not doing development, just want to build a commit:
  # Remember to update the hash / commit
  # src = fetchgit {
  #   rev = "commit hash here";
  #   url = "git@github.com:SkygearIO/skygear-server.git";
  #   sha256 = "hash here";
  # };
  buildInputs = [ git glide ]
    ++ (if withZMQ then [ libsodium zeromq pkgconfig ] else []);

  buildFlags = if withZMQ then "--tags zmq" else "";

  preBuild = ''
    pushd "go/src/${goPackagePath}"
    make vendor
    popd
  '';

  # Workaround a bug of the binary creates cycle reference
  # such that nix refuse to build.
  # https://github.com/NixOS/nixpkgs/issues/18131
  postInstall = if withZMQ then ''
    install_name_tool -delete_rpath $out/lib -add_rpath $bin $bin/bin/skygear-server
  '' else "";

  # Need this for glide to download dependencies
  SSL_CERT_FILE="${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt";

  meta = {
    homepage = "https://skygear.io";
    license = stdenv.lib.licenses.asl20;
  };
}
