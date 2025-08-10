#!/usr/bin/env bash

source "share/bash/completions/cli.bash"

cmdline="$@"

export COMP_LINE="${cmdline}"
export COMP_WORDS=(${cmdline})
export COMP_CWORD=$((${#COMP_WORDS[@]} - 1))
export COMP_POINT="${#cmdline}"

if [[ "$COMP_CWORD" -lt 0 ]]; then
  COMP_CWORD=0
fi

if [[ "${COMP_LINE}" == *" " ]]; then
  COMP_WORDS=(${COMP_WORDS[*]} "")
  COMP_CWORD=$((COMP_CWORD + 1))
fi

$(complete -p cli | sed "s/.*-F \\([^ ]*\\) .*/\\1/")
for comp in ${COMPREPLY[@]}; do
  echo $comp
done
