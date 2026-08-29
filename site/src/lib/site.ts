export const siteUrl = 'https://jedipunkz.rocks';
export const repoUrl = 'https://github.com/jedipunkz/fuzz.fish';
export const title = 'fuzz.fish - Fuzzy finder plugin for Fish Shell';
export const description =
  'fuzz.fish is a Fish Shell plugin for fuzzy finding command history, project files, Git branches, and commits from a unified terminal interface.';

export function absoluteUrl(path = '') {
  const base = import.meta.env.BASE_URL;

  return `${siteUrl}${base}${path}`;
}
