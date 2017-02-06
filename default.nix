let pkgs = import <nixpkgs> {};
in

{ stdenv         ? pkgs.stdenv
, buildGoPackage ? pkgs.buildGoPackage
, fetchgit       ? pkgs.fetchgit
, libsodium      ? pkgs.libsodium
, zeromq         ? pkgs.zeromq
, czmq           ? pkgs.czmq
, glide          ? pkgs.glide
, git            ? pkgs.git }:

buildGoPackage rec {
  name = "skygear";

  goPackagePath = "github.com/skygeario/skygear-server";

  # If you are doing development and want to build the current source:
  src = ./.;
  allowGoReference = true;

  # If you are not doing development, just want to build a commit:
  # Remember to update the hash / commit
  # src = fetchgit {
  #   rev = "commit hash here";
  #   url = "git@github.com:SkygearIO/skygear-server.git";
  #   sha256 = "hash here";
  # };
  buildInputs = [ libsodium zeromq czmq git glide ];

  preBuild = ''
    pushd "go/src/${goPackagePath}"
    make vendor
    popd
  '';

  # Need this for glide to download dependencies
  SSL_CERT_FILE="${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt";

  meta = {
    homepage = "https://skygear.io";
    license = stdenv.lib.licenses.asl20;
  };
}
