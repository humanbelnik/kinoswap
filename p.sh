#!/bin/bash

PROJECT_ROOT="$(git rev-parse --show-toplevel)"
cd "$PROJECT_ROOT" || exit 1

declare -a MAKE_TARGETS=("pre-commit")


for target in "${MAKE_TARGETS[@]}"; do
    if ! make "$target"; then
    echo "pre-commit hook failed"
    exit 1
    fi
done

echo "pre-commit hook pass"
exit 0