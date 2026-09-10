{
  description = "Ping TCP ports using tcping. Inspired by Linux's ping utility. Written in Go";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});

      # Read from the version package, the same file the Makefile reads, so a
      # release bump never has to be repeated here.
      version = builtins.elemAt (builtins.match ''.*var Current = "([^"]*)".*'' (
        builtins.readFile ./internal/version/version.go
      )) 0;
    in
    {
      packages = forAllSystems (pkgs: rec {
        default = tcping;

        tcping = pkgs.buildGoModule {
          pname = "tcping";
          inherit version;

          src = ./.;

          # Hash of the downloaded modules. It changes whenever go.sum does,
          # and "make nix-update" is what refreshes it.
          vendorHash = "sha256-x7HvBlAUyl9mJEJ6R40syjKjxGUhZzwXuq1rtucLhJ4=";

          # Same flags as the Makefile, so the Nix build produces the same
          # static binary the release archives carry.
          env.CGO_ENABLED = 0;
          ldflags = [
            "-s"
            "-w"
          ];

          subPackages = [ "cmd/tcping" ];

          postInstall = ''
            install -Dm644 completions/tcping.bash $out/share/bash-completion/completions/tcping
            install -Dm644 completions/_tcping $out/share/zsh/site-functions/_tcping
            install -Dm644 completions/tcping.fish $out/share/fish/vendor_completions.d/tcping.fish
          '';

          meta = {
            description = "Ping TCP ports using tcping. Inspired by Linux's ping utility. Written in Go";
            homepage = "https://github.com/pouriyajamshidi/tcping";
            license = nixpkgs.lib.licenses.mit;
            mainProgram = "tcping";
          };
        };
      });

      # Everything "make check" and "make release" need, pinned to the same
      # nixpkgs as the build above.
      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gnumake
            zip
            vhs
            tmux
          ];
        };
      });
    };
}
