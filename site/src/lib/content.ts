export const links = {
  repo: 'https://github.com/jedipunkz/fuzz.fish',
  releases: 'https://github.com/jedipunkz/fuzz.fish/releases',
  issues: 'https://github.com/jedipunkz/fuzz.fish/issues',
  pulls: 'https://github.com/jedipunkz/fuzz.fish/pulls',
  license: 'https://github.com/jedipunkz/fuzz.fish/blob/main/LICENSE',
  readme: 'https://github.com/jedipunkz/fuzz.fish#readme',
  fisher: 'https://github.com/jorgebucaran/fisher',
  fish: 'https://fishshell.com/',
};

export const installCommand = 'fisher install jedipunkz/fuzz.fish';

export const facts = [
  'MIT licensed',
  'Single Go binary',
  'No fzf required',
  'Fish 3.0+',
];

export type Mode = {
  key: string;
  name: string;
  title: string;
  text: string;
  accent: 'blue' | 'magenta' | 'green' | 'cyan' | 'orange';
};

export const modes: Mode[] = [
  {
    key: 'ctrl+r',
    name: 'History',
    title: 'Command history, ranked by frecency',
    text: 'Search the Fish history file with fuzzy matching. Recent and frequently used commands rise to the top, and enter puts the command back on the prompt instead of running it blind.',
    accent: 'blue',
  },
  {
    key: 'ctrl+s',
    name: 'Files',
    title: 'Files and directories from here',
    text: 'Walk the current directory while skipping hidden paths and build output such as node_modules and vendor. Insert a file path, or cd straight into a directory.',
    accent: 'green',
  },
  {
    key: 'ctrl+g',
    name: 'Branches',
    title: 'Switch branches without leaving the prompt',
    text: 'List local branches with their latest commit. Enter switches, and pressing ctrl+g again on the current branch runs git pull origin for it.',
    accent: 'magenta',
  },
  {
    key: 'ctrl+w',
    name: 'Worktrees',
    title: 'Jump between Git worktrees',
    text: 'Search worktrees by path, see the branch checked out in each one, and cd into the selected worktree.',
    accent: 'cyan',
  },
  {
    key: 'ctrl+x',
    name: 'Commits',
    title: 'Find a commit, pick the command',
    text: 'Match on short hash or subject and preview the diffstat. Enter opens a small action list — git show, diff, revert, cherry-pick, rebase --onto — and places the command on the prompt unrun.',
    accent: 'orange',
  },
];

export type KeyRow = {
  keys: string[];
  action: string;
};

export const modeKeys: KeyRow[] = [
  { keys: ['ctrl+r'], action: 'Command history search (default mode)' },
  { keys: ['ctrl+s'], action: 'File search' },
  { keys: ['ctrl+g'], action: 'Git branch search' },
  { keys: ['ctrl+w'], action: 'Git worktree search' },
  { keys: ['ctrl+x'], action: 'Git commit search' },
];

export const commonKeys: KeyRow[] = [
  { keys: ['↑ ↓', 'ctrl+p ctrl+n'], action: 'Move the selection' },
  { keys: ['tab'], action: 'Complete the query with the selected item' },
  { keys: ['enter'], action: 'Accept the selection for the current mode' },
  { keys: ['ctrl+y'], action: 'Copy the selected item to the clipboard' },
  { keys: ['esc', 'ctrl+c'], action: 'Cancel and return to the shell' },
];

export const details = [
  {
    title: 'Your prompt is the query',
    text: 'Whatever is already typed pre-fills the search box. Type vim, press ctrl+r, and history is narrowed to vim before you touch another key.',
  },
  {
    title: 'A * switches to glob',
    text: 'Any query containing * matches as a glob in every mode: nvim *.go finds nvim internal/app/filter.go, not commands that merely share those letters.',
  },
  {
    title: 'One binary, no finder to install',
    text: 'The TUI is a Go program built on Bubble Tea and shipped with the plugin. There is no fzf, no Python, and no runtime to keep in sync.',
  },
];

export const steps = [
  {
    label: 'Install Fisher',
    command: 'curl -sL https://raw.githubusercontent.com/jorgebucaran/fisher/main/functions/fisher.fish | source && fisher install jorgebucaran/fisher',
    note: 'Skip this if Fisher is already set up.',
  },
  {
    label: 'Install the plugin',
    command: installCommand,
    note: 'The Go binary is built once during install.',
  },
  {
    label: 'Press ctrl+r',
    command: null,
    note: 'History search opens immediately. Switch modes with a single key.',
  },
];

export const maintenance: [string, string][] = [
  ['Update', 'fisher update jedipunkz/fuzz.fish'],
  ['Uninstall', 'fisher remove jedipunkz/fuzz.fish'],
];

export const requirements = [
  ['Fish Shell 3.0+', 'The plugin installs into $__fish_config_dir.'],
  ['Go 1.24+', 'Used once to build the bundled binary.'],
  ['Git', 'Only for the branch, worktree, and commit modes.'],
];

export const contributing = [
  {
    title: 'Report an issue',
    text: 'Bugs, rough edges, and mode ideas are all welcome. Include your Fish and Go versions.',
    href: links.issues,
    cta: 'Open an issue',
  },
  {
    title: 'Send a pull request',
    text: 'CI runs go build, go vet, go test, and golangci-lint on every push and pull request to main.',
    href: links.pulls,
    cta: 'Browse pull requests',
  },
  {
    title: 'Read the source',
    text: 'A small Go codebase: a Bubble Tea app in internal/app, plus history, git, files, and scoring packages.',
    href: links.repo,
    cta: 'View the repository',
  },
];
