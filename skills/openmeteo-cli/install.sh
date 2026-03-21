#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 1 ]; then
  printf 'usage: %s /path/to/clawbot/skills\n' "$0" >&2
  exit 1
fi

target_root="$1"
script_dir="$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)"
target_dir="${target_root%/}/openmeteo-cli"
repo_slug="${REPO_SLUG:-ksinistr/openmeteo-cli}"
ref="${VERSION:-main}"
skill_path="skills/openmeteo-cli/SKILL.md"

mkdir -p "${target_dir}"

if [ -f "${script_dir}/SKILL.md" ]; then
  cp "${script_dir}/SKILL.md" "${target_dir}/SKILL.md"
elif [ -n "${RAW_BASE_URL:-}" ]; then
  curl -fsSL "${RAW_BASE_URL}/SKILL.md" -o "${target_dir}/SKILL.md"
else
  curl -fsSL "https://raw.githubusercontent.com/${repo_slug}/${ref}/${skill_path}" -o "${target_dir}/SKILL.md"
fi

printf 'installed openmeteo-cli skill to %s\n' "${target_dir}"
