#!/usr/bin/env fish

set -l sdir (dirname (realpath (status --current-filename)))
set -l cgen $argv[1]

pushd (mktemp -d)

function check_presence
    if test "$expected" = "~"
        return 0
    end
    for e in $expected
        if not contains -- $e $completions
            return 1
        end
    end
    for c in $completions
        if not contains -- $c $expected
            return 1
        end
    end
    return 0
end

set successful 0
set failed 0

for case in $sdir/case*
    if test -n "$cgen"
        $cgen $case/cli.yml
    else if test -f $sdir/../build/cgen
        $sdir/../build/cgen $case/cli.yml
    else
        cgen $case/cli.yml
    end

    set -l desc (cat $case/desc.txt)
    set -l expected_bash (cat $case/bash.output)
    set -l expected_fish (cat $case/fish.output)
    set -l expected_zsh (cat $case/zsh.output)
    set -l input (cat $case/input.txt)

    printf "Running test case $(basename $case): $desc\n"
    set -l errors false

    set_color yellow
    printf "\t[    RUN] BASH\n"
    set_color normal
    set -g completions (bash $sdir/test.bash "$input")
    set -g expected $expected_bash
    check_presence
    if test $status -eq 1
        set failed (math $failed + 1)
        set_color red
        printf "\t[FAILED ] BASH returned [$completions] instead of [$expected_bash]\n"
        set errors true
    else
        set successful (math $successful + 1)
        set_color green
        printf "\t[OK     ] BASH test finished successfully\n"
    end
    set_color normal

    set_color yellow
    printf "\t[    RUN] Fish\n"
    set_color normal
    set -g completions (fish $sdir/test.fish "$input")
    set -g expected $expected_fish
    check_presence
    if test $status -eq 1
        set failed (math $failed + 1)
        set_color red
        printf "\t[FAILED] Fish returned [$completions] instead of [$expected_fish]\n"
        set errors true
    else
        set successful (math $successful + 1)
        set_color green
        printf "\t[OK     ] Fish test finished successfully\n"
    end
    set_color normal

    set_color yellow
    printf "\t[    RUN] ZSH\n"
    set_color normal
    set -g completions (zsh $sdir/test.zsh -p "$input")
    set -g expected $expected_zsh
    check_presence
    if test $status -eq 1
        set failed (math $failed + 1)
        set_color red
        printf "\t[FAILED ] ZSH returned [$completions] instead of [$expected_zsh]\n"
        set errors true
    else
        set successful (math $successful + 1)
        set_color green
        printf "\t[OK     ] ZSH test finished successfully\n"
    end
    set_color normal

    if ! $errors
        set_color green
        printf "\t[SUCCESS]\n"
    else
        set_color red
        printf "\t[FAILED ]\n"
    end
    set_color normal

    rm -rf share
end

echo "$successful tests finished successfully, $failed failed"

popd
