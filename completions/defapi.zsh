#compdef defapi

_defapi() {
    local state

    _arguments \
        '1: :->category' \
        '2: :->model' \
        '*: :->args'

    case $state in
        category)
            _values 'category' \
                'text[Text generation]' \
                'image[Image generation]' \
                'video[Video generation]'
            ;;
        model)
            case ${words[2]} in
                text)
                    _values 'model' \
                        'gpt[OpenAI GPT-5.6 Sol]' \
                        'openai[OpenAI GPT-5.6 Sol]' \
                        'gemini[Google Gemini 3.8 Flash]' \
                        'google[Google Gemini 3.8 Flash]' \
                        'claude[Anthropic Claude Opus 5]' \
                        'anthropic[Anthropic Claude Opus 5]'
                    ;;
                image)
                    _values 'model' \
                        'wan[Alibaba Wan 2.7 Image]' \
                        'mj[Midjourney]' \
                        'gpt[OpenAI GPT-Image-1.5/2]' \
                        'gpt2[OpenAI GPT-Image-2]' \
                        'google[Google image models]'
                    ;;
                video)
                    _values 'model' \
                        'seedance[ByteDance Seedance 2.0]' \
                        'grok[xAI Grok Imagine Video]' \
                        'sora[OpenAI Sora 2 Stable]' \
                        'kling[Kuaishou Kling 4.0]'
                    ;;
            esac
            ;;
        args)
            case ${words[2]} in
                text)
                    case ${words[3]} in
                        gpt|openai)
                            _arguments \
                                '--model[Model]:model:(openai/gpt-5.6-sol openai/gpt-5.2)' \
                                '--system[System instruction]:system:' \
                                '--stream[Stream response]' \
                                '--max-tokens[Maximum output tokens]:tokens:' \
                                '--temperature[Sampling temperature]:temperature:' \
                                '--top-p[Nucleus sampling value]:top_p:' \
                                '(-o --output)'{-o,--output}'[Save generated text to this file path]:file:_files' \
                                '*:prompt:'
                            ;;
                        gemini|google)
                            _arguments \
                                '--model[Model]:model:(gemini-3.8-flash google/gemini-3.8-flash gemini-3.1-pro-preview google/gemini-3.1-pro-preview)' \
                                '--system[System instruction]:system:' \
                                '--stream[Stream response]' \
                                '--max-tokens[Maximum output tokens]:tokens:' \
                                '--temperature[Sampling temperature]:temperature:' \
                                '--top-p[Nucleus sampling value]:top_p:' \
                                '(-o --output)'{-o,--output}'[Save generated text to this file path]:file:_files' \
                                '*:prompt:'
                            ;;
                        claude|anthropic)
                            _arguments \
                                '--model[Model]:model:(anthropic/claude-opus-5 anthropic/claude-sonnet-4.6)' \
                                '--system[System instruction]:system:' \
                                '--stream[Stream response]' \
                                '--max-tokens[Maximum output tokens]:tokens:' \
                                '--temperature[Sampling temperature]:temperature:' \
                                '--top-p[Nucleus sampling value]:top_p:' \
                                '(-o --output)'{-o,--output}'[Save generated text to this file path]:file:_files' \
                                '*:prompt:'
                            ;;
                    esac
                    ;;
                image)
                    case ${words[3]} in
                        wan)
                            _arguments \
                                '--ratio[Aspect ratio]:ratio:(1:1 16:9 4:3 21:9 3:4 9:16 8:1)' \
                                '(-o --output)'{-o,--output}'[Save image to this file path]:file:_files' \
                                '--open[Open image after download]' \
                                '*:prompt:'
                            ;;
                        mj|midjourney)
                            _arguments \
                                '--speed[Processing speed]:speed:(fast relax)' \
                                '--bot[Bot type]:bot:(MID_JOURNEY NIJI_JOURNEY)' \
                                '--image[Image URL for editing]:url:' \
                                '(-o --output)'{-o,--output}'[Save image to this file path]:file:_files' \
                                '--open[Open image after download]' \
                                '*:prompt:'
                            ;;
                        gpt|gpt2)
                            _arguments \
                                '--model[Model]:model:(gpt-image-1.5 gpt-image-2)' \
                                '--size[Output size]:size:(auto 1024x1024 1536x1024 1024x1536 1:1 16:9 9:16)' \
                                '--quality[Quality]:quality:(auto high medium low)' \
                                '--background[Background]:background:(auto opaque transparent)' \
                                '--format[Output format]:format:(png jpeg webp)' \
                                '--image[Reference image URL for gpt-image-2]:url:' \
                                '(-o --output)'{-o,--output}'[Save image to this file path]:file:_files' \
                                '--open[Open image after download]' \
                                '*:prompt:'
                            ;;
                        google)
                            _arguments \
                                '--model[Model]:model:(nano-banana nano-banana-pro nano-banana-2 gemini-2.5-flash-image gemini-3.1-flash-image-preview)' \
                                '--ratio[Aspect ratio]:ratio:(auto 1:1 16:9 21:9 2:3 3:2 3:4 4:3 4:5 5:4 9:16)' \
                                '--size[Output resolution]:size:(1k 2k 4k)' \
                                '(-o --output)'{-o,--output}'[Save image to this file path]:file:_files' \
                                '--open[Open image after download]' \
                                '*:prompt:'
                            ;;
                    esac
                    ;;
                video)
                    case ${words[3]} in
                        seedance)
                            _arguments \
                                '--duration[Duration in seconds]:duration:(5 10 15)' \
                                '--ratio[Aspect ratio]:ratio:(16:9 9:16 1:1 4:3 3:4 21:9)' \
                                '--image[Reference image URL]:url:' \
                                '--open[Open video after download]' \
                                '*:prompt:'
                            ;;
                        grok)
                            _arguments \
                                '--duration[Duration in seconds]:duration:(10 15)' \
                                '--ratio[Aspect ratio]:ratio:(16:9 9:16 1:1 2:3 3:2)' \
                                '--image[Reference image URL]:url:' \
                                '--open[Open video after download]' \
                                '*:prompt:'
                            ;;
                        sora)
                            _arguments \
                                '--duration[Duration in seconds]:duration:(10 15 25)' \
                                '--ratio[Aspect ratio]:ratio:(16:9 9:16)' \
                                '--variant[Model variant]:variant:(sora-2 sora-2-hd sora-2-pro)' \
                                '--image[Reference image URL]:url:' \
                                '--open[Open video after download]' \
                                '*:prompt:'
                            ;;
                        kling)
                            _arguments \
                                '--duration[Duration in seconds (3-30)]:duration:(3 5 10 15 20 30)' \
                                '--resolution[Resolution]:resolution:(720p 1080p)' \
                                '--ratio[Aspect ratio]:ratio:(auto 16:9 9:16 1:1 21:9)' \
                                '*--image[Reference image URL]:url:' \
                                '*--video[Reference video URL]:url:' \
                                '--keep-sound[Keep reference video audio]' \
                                '--first-frame[First frame image URL]:url:' \
                                '*--keyframe[Intermediate keyframe image URL]:url:' \
                                '--last-frame[Last frame image URL]:url:' \
                                '--open[Open video after download]' \
                                '*:prompt:'
                            ;;
                    esac
                    ;;
            esac
            ;;
    esac
}

_defapi
