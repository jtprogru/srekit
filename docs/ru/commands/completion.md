# srekit completion

Сгенерировать shell-автодополнение. Предоставляется cobra; поддерживает
`bash`, `zsh`, `fish`, `powershell`.

## Синопсис

```bash
srekit completion bash
srekit completion zsh
srekit completion fish
srekit completion powershell
```

Каждая подкоманда печатает скрипт в stdout; перенаправляешь в нужное
место для своего шелла.

## Примеры

=== "zsh"

    ```bash
    srekit completion zsh > "${fpath[1]}/_srekit"
    # restart shell или: autoload -Uz compinit && compinit
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
    # или дописать в существующий $PROFILE
    ```

## См. также

- [Cobra completion docs](https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md) —
  детали underlying механизма.
