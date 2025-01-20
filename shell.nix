# SPDX-FileCopyrightText: 2025 Henry Bubert
#
# SPDX-License-Identifier: MIT

{
  pkgs,
  pre-commit-hooks ? pkgs.pre-commit-hooks,
}: let
  pre-commit-check = pre-commit-hooks.lib.${pkgs.system}.run {
    src = ./.;
    hooks = {
      gotest.enable = true;
      gofmt.enable = true;
      #revive.enable = true;
      goimports = {
        enable = true;
        name = "goimports";
        description = "Format my golang code";
        files = "\.go$";
        entry = let
          script = pkgs.writeShellScript "precommit-goimports" ''
            set -e
            failed=false
            for file in "$@"; do
                # redirect stderr so that violations and summaries are properly interleaved.
                if ! ${pkgs.gotools}/bin/goimports -l -d "$file" 2>&1
                then
                    failed=true
                fi
            done
            if [[ $failed == "true" ]]; then
                exit 1
            fi
          '';
        in
          builtins.toString script;
      };
    };
  };
in
  pkgs.mkShell {
    packages = with pkgs; [
        # handy
        jq
        reuse

        # dev tools
        go_1_22
        go-outline
        gopls
        gopkgs
        go-tools
        delve
        revive
        errcheck
        unconvert
        godef
        clang

        # mass deps
        postgresql
        gotools # for stringer
    ];

    shellHook =
      pre-commit-check.shellHook
      + ''
        env_up

        export DBPATH=$PWD/tmp/db
        isNewPGInstance=0
        if ! test -d ./tmp/db; then
          # Initialize PostgreSQL instance
          initdb -D $DBPATH
          isNewPGInstance=1
        fi

        export DATABASE_URL="postgres://localhost:5432/pgmmr-test"
        export PGHOST=$PWD/tmp
        # Check if PostgreSQL instance is already running
        if ! pg_isready >/dev/null 2>&1; then
          pg_ctl -D $DBPATH -l $PWD/tmp/pglogfile -o "--unix_socket_directories='$PWD/tmp'" start
        fi

        export PGDATABASE=$(echo $DATABASE_URL | cut -d'/' -f4 | cut -d'?' -f1)
        # TODO check if database exists
        if [ "$isNewPGInstance" -eq "1" ]; then
          createdb pgmmr-test
          psql pgmmr-test < ./test_schema.sql
        fi

        # shutdown postgres and ipfs when we exit
        cleanup() {
            pg_ctl -D $DBPATH stop
        }
        trap cleanup EXIT
      '';
  }
