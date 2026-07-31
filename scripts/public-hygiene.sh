#!/usr/bin/env bash
set -Eeuo pipefail

failures=()

while IFS= read -r tracked_path; do
  [[ -e "${tracked_path}" || -L "${tracked_path}" ]] || continue
  case "${tracked_path}" in
    .codex/*|.claude/*|.agents/*|CLAUDE.md)
      failures+=("local agent configuration is tracked: ${tracked_path}")
      ;;
    .env|.env.*)
      if [[ "${tracked_path}" != ".env.example" ]]; then
        failures+=("environment file is tracked: ${tracked_path}")
      fi
      ;;
    *.pem|*.key|*.p12|*.pfx|*.sqlite|*.sqlite-shm|*.sqlite-wal|*.dump|*.log)
      failures+=("sensitive or runtime artifact is tracked: ${tracked_path}")
      ;;
  esac
done < <(git ls-files)

if ((${#failures[@]} > 0)); then
  printf -- '- %s\n' "${failures[@]}" >&2
  exit 1
fi

printf 'Public repository hygiene passed.\n'
