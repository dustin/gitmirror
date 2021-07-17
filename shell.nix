with import <nixpkgs> {};

stdenv.mkDerivation {
  name = "haskell";
  buildInputs = [
    go
  ];
}
