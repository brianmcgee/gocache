{inputs, ...}: {
  imports = [
    inputs.flake-root.flakeModule
    ./formatter.nix
    ./shell.nix
  ];
}
