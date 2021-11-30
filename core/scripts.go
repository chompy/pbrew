package core

const brewInstall = `
cd %s
curl -L https://github.com/Homebrew/brew/archive/refs/tags/3.3.6.tar.gz | tar xz --strip 1
`
