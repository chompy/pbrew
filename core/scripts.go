package core

const brewInstall = `
cd %s
curl -L https://github.com/Homebrew/brew/tarball/master | tar xz --strip 1
`
