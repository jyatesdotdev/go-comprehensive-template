# AGENTS.md — internal/cli

Reusable CLI helpers: configuration loading (Viper) and output formatting (table / JSON / YAML,
with ANSI color). Consumed by `examples/cli` (a Cobra app).

Read [internal/AGENTS.md](../AGENTS.md) and the [root](../../AGENTS.md) first.

## Files

- `config.go` — `Config` (+ `AppConfig`, `DatabaseConfig`, `LogConfig`), `LoadConfig`, `ConfigFileUsed`.
- `output.go` — `OutputFormat`, `ParseFormat`, `Printer`, `Colorize`, ANSI color consts, status printers.
- `cli_test.go` — tests for parsing/formatting/loading.

**This is the one `internal/` package with third-party production deps:**
`github.com/spf13/viper` and `gopkg.in/yaml.v3` (Cobra lives in `examples/cli`, not here).

## Config loading (`config.go`)

`LoadConfig(cfgFile, envPrefix)` resolves configuration with this precedence and rules:

- **Sources:** explicit `cfgFile` if set, otherwise a `config.yaml` searched in `.` and
  `$HOME/.myapp`. Then environment variables, then built-in defaults.
- **Env binding:** `SetEnvPrefix(envPrefix)` + `.`→`_` replacer + `AutomaticEnv`, so
  `app.port` ← `MYAPP_APP_PORT` (when `envPrefix="MYAPP"`).
- **Defaults:** `app.name=myapp`, `app.port=8080`, `database.host=localhost`,
  `database.port=5432`, `log.level=info`, `log.format=text`. Keep struct fields, `mapstructure`
  tags, and defaults in sync when you add a setting.
- **Missing config file is not an error** — `ConfigFileNotFoundError` is swallowed so defaults +
  env still apply. Any other read error is wrapped and returned.

> Note: `LoadConfig` uses the Viper singleton (package-level `viper.*`), so it is process-global
> and not suited to concurrent multi-config loading. Fine for a CLI; be aware if reusing elsewhere.

## Output formatting (`output.go`)

- **`OutputFormat`** ∈ `table` | `json` | `yaml`. `ParseFormat` **defaults to `table`** for any
  unrecognized string (never errors).
- **`Printer{ Out, Format, NoColor }`** — `NewPrinter` writes to `os.Stdout` by default; callers
  often reassign `Out` (e.g. to `cmd.OutOrStdout()` for testability).
- **`Print(data)` contract:** JSON/YAML accept any value; **table format requires `[][]string`
  where row 0 is the header** — anything else returns an error (`"table format requires [][]string"`).
  Convert structs to rows before printing tables (see `examples/cli`'s `list`).
- **Color:** `Colorize(color, text, noColor)` returns plain text when `noColor` is true. Status
  helpers `Success`/`Error`/`Info`/`Warn` prefix `✓/✗/ℹ/⚠` and color accordingly. Always thread
  `NoColor` through so `--no-color` and non-TTY output stay clean.

## Go best practices for CLI code

- Keep flag/command wiring in the `main` package (`examples/cli`); keep this package free of
  global command state so it stays reusable and testable.
- Write to an injected `io.Writer`, not directly to `os.Stdout`, so output can be captured in tests.
- Bind flags→Viper for layered config, but keep binding failures non-fatal
  (`// #nosec G104 -- bind failure is non-fatal`).

## Testing

Redirect `Printer.Out` to a `bytes.Buffer` and assert rendered output. For config, set env vars
or point `cfgFile` at a temp file. Table output alignment comes from `text/tabwriter` — assert on
content, not exact spacing.
