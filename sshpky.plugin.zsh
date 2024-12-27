__MYSSHP_PY_DIR="${ZSH}/plugins/sshpky"

local cmd="$__MYSSHP_PY_DIR/sshpky.py"


function __complete_ssh_host() {
    local KNOWN_FILE=~/.ssh/known_hosts
    local CONFIG_FILE=~/.ssh/config
    local HOSTS_FILE=/etc/hosts

    if [ -r "$KNOWN_FILE" ]; then
        local KNOWN_LIST=$(cut -f 1 -d ' ' "$KNOWN_FILE" | cut -f 1 -d ',' | grep -v '^[0-9[]')
    fi

    if [ -r "$CONFIG_FILE" ]; then
        local CONFIG_LIST=$(awk '/^Host [A-Za-z0-9]+/ {print $2}' "$CONFIG_FILE")
    fi

    if [ -r "$HOSTS_FILE" ]; then
        local HOSTS_LIST=$(awk '/^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+/ {print $2}' "$HOSTS_FILE")
    fi

    local PARTIAL_WORD="${COMP_WORDS[COMP_CWORD]}"
    COMPREPLY=( $(compgen -W "$KNOWN_LIST$IFS$CONFIG_LIST$IFS$HOSTS_LIST" -- "$PARTIAL_WORD") )
    compadd -a COMPREPLY
}

# 替代别名的函数
function sshpky() {
    python3 $cmd "$@"
}

# alias sshpky='python3 ${cmd}'

compdef __complete_ssh_host sshpky

