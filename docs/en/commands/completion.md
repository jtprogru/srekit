# srekit completion

Generate shell autocomplete scripts. Provided by cobra; supports `bash`, `zsh`, `fish`, and `powershell`.

## Synopsis

```bash
srekit completion bash
srekit completion zsh
srekit completion fish
srekit completion powershell
```

Each subcommand emits the script to stdout; redirect into the right location for your shell.

## Examples

=== "zsh"

    ```bash
    srekit completion zsh > "${fpath[1]}/_srekit"
    # restart shell or: autoload -Uz compinit && compinit
    ```

=== "bash"

    System-wide (root):

    ```bash
    srekit completion bash > /etc/bash_completion.d/srekit
    ```

    Per-user:

    ```bash
    srekit completion bash > "${BASH_COMPLETION_USER_DIR:-$HOME/.local/share/bash-completion}/completions/srekit"
    ```

=== "fish"

    ```fish
    srekit completion fish > ~/.config/fish/completions/srekit.fish
    ```

=== "powershell"

    ```powershell
    srekit completion powershell > $PROFILE.CurrentUserAllHosts
    # or append to your existing $PROFILE
    ```

## See also

- [Cobra's completion docs](https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md) for details on the underlying mechanism.
