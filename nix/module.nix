self: {
  config,
  lib,
  pkgs,
  ...
}: let
  cfg = config.services.altar;
in {
  options.services.spendable = {
    enable = lib.mkEnableOption "the Altar service";

    package = lib.mkOption {
      type = lib.types.package;
      default = self.packages.${pkgs.stdenv.hostPlatform.system}.default;
      defaultText = lib.literalMD "the flake's `packages.<system>.default`";
      description = "The altar package to run.";
    };

    environmentFile = lib.mkOption {
      type = lib.types.path;
      example = "/run/secrets/altar.env";
      description = ''
        ```
        AWTRIX_IP=""
        LATITUDE="-37.814"
        LONGITUDE="144.9633"
        WEATHER_TIMEZONE="Australia/Sydney"
        STARGAZER_WEBHOOK_SECRET=""
        ```
      '';
    };
  };

  config = lib.mkIf cfg.enable {
    systemd.services.spendable = {
      description = "Awtrix Listens To Altar Requests";
      wantedBy = ["multi-user.target"];
      after = ["network-online.target"];
      wants = ["network-online.target"];

      serviceConfig = {
        ExecStart = lib.getExe cfg.package;
        EnvironmentFile = cfg.environmentFile;
        Restart = "on-failure";
        RestartSec = 5;

        # Hardening
        DynamicUser = true;
        NoNewPrivileges = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        PrivateTmp = true;
        PrivateDevices = true;
        ProtectKernelTunables = true;
        ProtectKernelModules = true;
        ProtectControlGroups = true;
        RestrictAddressFamilies = [
          "AF_INET"
          "AF_INET6"
        ];
        RestrictNamespaces = true;
        LockPersonality = true;
        MemoryDenyWriteExecute = true;
        SystemCallArchitectures = "native";
      };
    };
  };
}
