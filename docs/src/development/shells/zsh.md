# Zsh

## Case-insensitive matching

`CARAPACE_MATCH=CASE_INSENSITIVE` makes carapace return case-insensitive matches,
but zsh still filters the candidates it receives. Configure zsh's completion
matcher as well:

```zsh
zstyle ':completion:*' matcher-list 'm:{a-zA-Z}={A-Za-z}'
```

For example, add the matcher alongside the carapace completion script:

```zsh
export CARAPACE_MATCH=CASE_INSENSITIVE
zstyle ':completion:*' matcher-list 'm:{a-zA-Z}={A-Za-z}'
source <(command _carapace)
```

Without the `zstyle` matcher, zsh can filter out candidates that carapace
already matched.
