# CLI Development in Go

This guide covers building production-quality CLIs with Cobra and Viper, including patterns demonstrated in `examples/cli/` and `internal/cli/`.

## Cobra Fundamentals

Cobra organizes CLIs as a tree of commands. Each command can have subcommands, flags, and argument validation.

### Root Command

The root command is the entry point. Persistent flags defined here are inherited by all subcommands:

```go
var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "A demo CLI showing Cobra + Viper patterns",
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
        // Runs before any subcommand — good for shared setup
        printer = cli.NewPrinter(cli.ParseFormat(output), noColor)
        printer.Out = cmd.OutOrStdout()
    },
}

func init() {
    rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
    rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "output format: table, json, yaml")
    rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
    rootCmd.AddCommand(greetCmd, configCmd, listCmd)
}
```

### Subcommands

Each subcommand is a `*cobra.Command` with its own flags and `RunE` function:

```go
var greetCmd = &cobra.Command{
    Use:   "greet",
    Short: "Print a greeting",
    Args:  cobra.NoArgs,
    RunE: func(cmd *cobra.Command, args []string) error {
        printer.Success(fmt.Sprintf("%s, %s!", greetGreeting, greetName))
        return nil
    },
}

func init() {
    greetCmd.Flags().StringVarP(&greetName, "name", "n", "World", "who to greet")
}
```

### Flags: Persistent vs Local

| Type | Scope | Example |
|------|-------|---------|
| Persistent | All subcommands | `--config`, `--output` |
| Local | Single subcommand | `--name`, `--count` |

Use `PersistentFlags()` on the root command for global options. Use `Flags()` on individual subcommands for command-specific options.

### Argument Validation

Cobra provides built-in validators:

```go
Args: cobra.NoArgs           // rejects any positional args
Args: cobra.ExactArgs(1)     // requires exactly 1
Args: cobra.MaximumNArgs(1)  // allows 0 or 1
Args: cobra.MinimumNArgs(1)  // requires at least 1
```

Custom validation:

```go
Args: func(cmd *cobra.Command, args []string) error {
    for _, a := range args {
        if !isValid(a) {
            return fmt.Errorf("invalid argument: %s", a)
        }
    }
    return nil
},
```

## Viper Integration

Viper handles configuration from files, environment variables, and defaults. See `internal/cli/config.go`.

### Config Loading Order

Viper merges configuration with this precedence (highest first):

1. Explicit flag values
2. Environment variables
3. Config file values
4. Defaults

### Environment Variable Binding

```go
viper.SetEnvPrefix("MYAPP")                              // MYAPP_ prefix
viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))    // app.port → MYAPP_APP_PORT
viper.AutomaticEnv()
```

### Flag Binding

Bind Cobra flags to Viper so config files and env vars can set them:

```go
viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
```

### Config File Search

```go
viper.SetConfigName("config")
viper.SetConfigType("yaml")
viper.AddConfigPath(".")
viper.AddConfigPath("$HOME/.myapp")
```

### Defaults

```go
viper.SetDefault("app.port", 8080)
viper.SetDefault("log.level", "info")
```

### Unmarshaling

```go
var cfg Config
if err := viper.Unmarshal(&cfg); err != nil {
    return nil, fmt.Errorf("unmarshaling config: %w", err)
}
```

Use `mapstructure` tags on struct fields to control mapping.

## Output Formatting

The `internal/cli` package provides a `Printer` that supports table, JSON, and YAML output. See `internal/cli/output.go`.

### Multi-Format Output

```go
printer := cli.NewPrinter(cli.ParseFormat(output), noColor)

// JSON/YAML: pass any serializable value
printer.Print(myStruct)

// Table: pass [][]string with headers as first row
printer.Print([][]string{
    {"NAME", "STATUS"},
    {"api", "running"},
})
```

### Colored Status Messages

```go
printer.Success("Deployed successfully")  // green ✓
printer.Error("Connection failed")        // red ✗
printer.Info("Using config: app.yaml")    // blue ℹ
printer.Warn("No config file found")      // yellow ⚠
```

Always support `--no-color` for CI environments and piped output.

### ANSI Color Helpers

```go
cli.Colorize(cli.Green, "text", noColor)  // wraps in ANSI codes
cli.Colorize(cli.Bold, "HEADER", noColor) // bold text
```

## Shell Completion

Cobra generates completion scripts for bash, zsh, fish, and PowerShell automatically via the built-in `completion` subcommand:

```bash
# Bash
myapp completion bash > /etc/bash_completion.d/myapp

# Zsh
myapp completion zsh > "${fpath[1]}/_myapp"

# Fish
myapp completion fish > ~/.config/fish/completions/myapp.fish
```

No extra code is needed — Cobra derives completions from your command tree, flags, and `ValidArgs`.

### Custom Completions

For dynamic values, register a completion function:

```go
cmd.RegisterFlagCompletionFunc("region", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    return []string{"us-east-1", "us-west-2", "eu-west-1"}, cobra.ShellCompDirectiveNoFileComp
})
```

## Testing CLI Commands

### Testing Cobra Commands

Use `SetArgs`, `SetOut`, and `SetErr` to capture output without running a real process:

```go
func executeCommand(args ...string) (string, error) {
    var buf bytes.Buffer
    rootCmd.SetOut(&buf)
    rootCmd.SetErr(&buf)
    rootCmd.SetArgs(args)
    err := rootCmd.Execute()
    return buf.String(), err
}

func TestGreetDefault(t *testing.T) {
    out, err := executeCommand("greet", "--no-color")
    if err != nil {
        t.Fatal(err)
    }
    if !strings.Contains(out, "Hello, World!") {
        t.Errorf("unexpected output: %q", out)
    }
}
```

Key pattern: use `cmd.OutOrStdout()` in your commands (not `os.Stdout`) so tests can capture output.

### Testing Output Helpers

Create a `Printer` with a `bytes.Buffer` to test formatting in isolation:

```go
var buf bytes.Buffer
p := &cli.Printer{Out: &buf, Format: cli.FormatJSON, NoColor: true}
p.Print(data)
// assert on buf.String()
```

### Testing Error Cases

Verify argument validation and error paths:

```go
func TestListTooManyArgs(t *testing.T) {
    _, err := executeCommand("list", "a", "b")
    if err == nil {
        t.Error("list with 2 args should fail")
    }
}
```

## Cobra vs urfave/cli

| Feature | Cobra | urfave/cli |
|---------|-------|------------|
| Subcommand nesting | Deep nesting, natural tree structure | Supported but less ergonomic |
| Flag types | Persistent + local, rich type support | Global + command-level |
| Completion | Built-in for bash/zsh/fish/powershell | Requires manual setup |
| Config integration | First-class Viper support | Manual integration |
| Docs generation | Built-in man page and markdown generation | Not built-in |
| Adoption | kubectl, docker, hugo, gh | Many smaller CLIs |
| API style | Struct-based command definitions | Struct-based, slightly more verbose |
| Learning curve | Moderate — more concepts | Lower — simpler API |

**When to choose Cobra**: Complex CLIs with deep subcommand trees, config file support, or shell completion requirements. The Viper integration makes it the default choice for most Go CLIs.

**When to choose urfave/cli**: Simpler tools where Cobra's feature set is overkill. Lighter dependency tree.

## Best Practices

1. **Use `RunE` not `Run`** — return errors instead of calling `os.Exit` directly. This makes commands testable and composable.

2. **Write to `cmd.OutOrStdout()`** — never write directly to `os.Stdout` in command handlers. This enables test capture and output redirection.

3. **Support `--output` and `--no-color`** — machine-readable output (JSON/YAML) is essential for scripting. Disable color when output is piped.

4. **Validate args early** — use Cobra's `Args` validators to reject bad input before running business logic.

5. **Keep commands thin** — commands should parse input and call into library code. Business logic belongs in packages, not in `RunE` functions.

6. **Use persistent pre-run for shared setup** — `PersistentPreRun` on the root command is the right place for initializing loggers, printers, and config.

7. **Provide shell completion** — it's free with Cobra and significantly improves UX.

8. **Exit codes matter** — return distinct exit codes for different failure modes. Use `os.Exit(1)` only in `main()`.

## Running the Examples

```bash
# Run the CLI demo
go run ./examples/cli/ --help
go run ./examples/cli/ greet --name Go
go run ./examples/cli/ list --output json
go run ./examples/cli/ config

# Run tests
go test ./internal/cli/ ./examples/cli/
```

## Further Reading

- [Cobra Documentation](https://cobra.dev)
- [Viper Documentation](https://github.com/spf13/viper)
- [urfave/cli](https://github.com/urfave/cli)
- [12-Factor App: Config](https://12factor.net/config)

## See Also

- [EXTENDING.md](EXTENDING.md) — Adding new commands and packages
- [Best Practices](best-practices.md) — Go idioms and error handling
- [Third-Party Libraries](third-party.md) — Cobra, Viper, and dependency management
