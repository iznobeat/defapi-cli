_defapi() {
    local cur prev words cword
    _init_completion || return

    if [[ $cword -eq 1 ]]; then
        COMPREPLY=($(compgen -W "text image video" -- "$cur"))
        return
    fi

    case "${words[1]}" in
        text)
            if [[ $cword -eq 2 ]]; then
                COMPREPLY=($(compgen -W "gpt openai gemini google claude anthropic" -- "$cur"))
                return
            fi
            case "${words[2]}" in
                gpt|openai)
                    case "$prev" in
                        --model) COMPREPLY=($(compgen -W "openai/gpt-5.6-sol openai/gpt-5.2" -- "$cur")) ;;
                        --output|-o) COMPREPLY=($(compgen -f -- "$cur")) ;;
                        --max-tokens|--temperature|--top-p|--system) COMPREPLY=() ;;
                        *) COMPREPLY=($(compgen -W "--model --system --stream --stream=false --max-tokens --temperature --top-p --output -o" -- "$cur")) ;;
                    esac ;;
                gemini|google)
                    case "$prev" in
                        --model) COMPREPLY=($(compgen -W "gemini-3.8-flash google/gemini-3.8-flash gemini-3.1-pro-preview google/gemini-3.1-pro-preview" -- "$cur")) ;;
                        --output|-o) COMPREPLY=($(compgen -f -- "$cur")) ;;
                        --max-tokens|--temperature|--top-p|--system) COMPREPLY=() ;;
                        *) COMPREPLY=($(compgen -W "--model --system --stream --stream=false --max-tokens --temperature --top-p --output -o" -- "$cur")) ;;
                    esac ;;
                claude|anthropic)
                    case "$prev" in
                        --model) COMPREPLY=($(compgen -W "anthropic/claude-opus-5 anthropic/claude-sonnet-4.6" -- "$cur")) ;;
                        --output|-o) COMPREPLY=($(compgen -f -- "$cur")) ;;
                        --max-tokens|--temperature|--top-p|--system) COMPREPLY=() ;;
                        *) COMPREPLY=($(compgen -W "--model --system --stream --stream=false --max-tokens --temperature --top-p --output -o" -- "$cur")) ;;
                    esac ;;
            esac ;;

        image)
            if [[ $cword -eq 2 ]]; then
                COMPREPLY=($(compgen -W "wan mj gpt gpt2 google" -- "$cur"))
                return
            fi
            case "${words[2]}" in
                wan)
                    case "$prev" in
                        --ratio) COMPREPLY=($(compgen -W "1:1 16:9 4:3 21:9 3:4 9:16 8:1" -- "$cur")) ;;
                        --output|-o) COMPREPLY=($(compgen -f -- "$cur")) ;;
                        *) COMPREPLY=($(compgen -W "--ratio --output -o --open" -- "$cur")) ;;
                    esac ;;
                mj|midjourney)
                    case "$prev" in
                        --speed) COMPREPLY=($(compgen -W "fast relax" -- "$cur")) ;;
                        --bot)   COMPREPLY=($(compgen -W "MID_JOURNEY NIJI_JOURNEY" -- "$cur")) ;;
                        --output|-o) COMPREPLY=($(compgen -f -- "$cur")) ;;
                        *) COMPREPLY=($(compgen -W "--speed --bot --image --output -o --open" -- "$cur")) ;;
                    esac ;;
                gpt|gpt2)
                    case "$prev" in
                        --model)      COMPREPLY=($(compgen -W "gpt-image-1.5 gpt-image-2" -- "$cur")) ;;
                        --size)       COMPREPLY=($(compgen -W "auto 1024x1024 1536x1024 1024x1536 1:1 16:9 9:16" -- "$cur")) ;;
                        --quality)    COMPREPLY=($(compgen -W "auto high medium low" -- "$cur")) ;;
                        --background) COMPREPLY=($(compgen -W "auto opaque transparent" -- "$cur")) ;;
                        --format)     COMPREPLY=($(compgen -W "png jpeg webp" -- "$cur")) ;;
                        --output|-o)  COMPREPLY=($(compgen -f -- "$cur")) ;;
                        *) COMPREPLY=($(compgen -W "--model --size --quality --background --format --image --output -o --open" -- "$cur")) ;;
                    esac ;;
                google)
                    case "$prev" in
                        --model) COMPREPLY=($(compgen -W "nano-banana nano-banana-pro nano-banana-2 gemini-2.5-flash-image gemini-3.1-flash-image-preview" -- "$cur")) ;;
                        --ratio) COMPREPLY=($(compgen -W "auto 1:1 16:9 21:9 2:3 3:2 3:4 4:3 4:5 5:4 9:16" -- "$cur")) ;;
                        --size)  COMPREPLY=($(compgen -W "1k 2k 4k" -- "$cur")) ;;
                        --output|-o) COMPREPLY=($(compgen -f -- "$cur")) ;;
                        *) COMPREPLY=($(compgen -W "--model --ratio --size --output -o --open" -- "$cur")) ;;
                    esac ;;
            esac ;;

        video)
            if [[ $cword -eq 2 ]]; then
                COMPREPLY=($(compgen -W "seedance grok sora" -- "$cur"))
                return
            fi
            case "${words[2]}" in
                seedance)
                    case "$prev" in
                        --duration) COMPREPLY=($(compgen -W "5 10 15" -- "$cur")) ;;
                        --ratio)    COMPREPLY=($(compgen -W "16:9 9:16 1:1 4:3 3:4 21:9" -- "$cur")) ;;
                        *) COMPREPLY=($(compgen -W "--duration --ratio --image --open" -- "$cur")) ;;
                    esac ;;
                grok)
                    case "$prev" in
                        --duration) COMPREPLY=($(compgen -W "10 15" -- "$cur")) ;;
                        --ratio)    COMPREPLY=($(compgen -W "16:9 9:16 1:1 2:3 3:2" -- "$cur")) ;;
                        *) COMPREPLY=($(compgen -W "--duration --ratio --image --open" -- "$cur")) ;;
                    esac ;;
                sora)
                    case "$prev" in
                        --duration) COMPREPLY=($(compgen -W "10 15 25" -- "$cur")) ;;
                        --ratio)    COMPREPLY=($(compgen -W "16:9 9:16" -- "$cur")) ;;
                        --variant)  COMPREPLY=($(compgen -W "sora-2 sora-2-hd sora-2-pro" -- "$cur")) ;;
                        *) COMPREPLY=($(compgen -W "--duration --ratio --variant --image --open" -- "$cur")) ;;
                    esac ;;
            esac ;;
    esac
}

complete -F _defapi defapi
