# sophrosyne

## Todo

1. [ ] Add a `--help` option to the CLI.
1. [ ] Add a `--version` option to the CLI.
1. [ ] Add a `config` command to the CLI.
1. [ ] Add integration tests for the `config` command.
1. [ ] Add integration tests for the `healthcheck` command.
1. [ ] Add integration tests for the `--help` option.
1. [ ] Add integration tests for the `--version` option.
1. [ ] 100% unit test coverage.
1. [ ] Add integration tests for the checks API endpoints
1. [ ] Add integration tests for the profiles API endpoints
1. [ ] Add integration tests for the safety API endpoints
1. [ ] Implement authorization layer via Cedar
1. [ ] Add `migration` command to the CLI to migrate the database using alembic
   instead of running `alembic upgrade head` manually. This would also allow
    us to get the database URL from the configuration file.
1. [ ] Add integration tests for the `migration` command.
1. [ ] Properly document that the `security.site_key` and `security.salt`
    should be kept secret and cannot be changed after the first run.
1. [ ] Determine if there is a way to rotate the `security.site_key` and
    `security.salt` without breaking the system.
1. [ ] If `security.site_key` and `security.salt` are not set in the
    configuration file, if they are empty or not long enough, the application
    should complain at start and exit.
1. [ ] Implement Audit Logging
1. [ ] Implement proper error handling
    1. [ ] Error responses should be JSON
1. [ ] Allow specifying a profile to use when requesting a scan
