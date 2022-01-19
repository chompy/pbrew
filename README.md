PBREW
=====
**By Contextual Code**

A tool for quickly setting up local Platform.sh development environments that uses Homebrew and doesn't require containers.


## Installation

Currently precompiled binaries are not available for PBREW. You can build it yourself with Go.

1. Download and install Go, https://go.dev/dl/
2. Clone this repo.
3. Run `go build`
4. If desired copy, symlink, or create an alias to the pbrew binary.

Running PBREW for the first time will create a Homebrew environment in `~/.pbrew/homebrew`.


## Supported Services
PBREW is designed to support the services we use at Contextual Code. The following are the services it supports...

- PHP 5.6, 7.0, 7.1, 7.2, 7.3, 7.4
- MariaDB 10.6
- Redis 6.2
- SOLR 7.7


## Usage

In the root of a project you can run `pbrew p:start` to start PBREW. When ran for the first time all the nessacary services will be installed, it can take quite a long time.

### Pre-Install All Services
You can pre-install all of PBREW's services with the `pbrew brew:install-all` command. This is good if you want to leave your computer on over night to get everything setup.

### Build/Deploy/Application Dependencies
Currently PBREW does not automatically run build or deploy hooks. You have to run them manually.

```
pbrew app:build
pbrew app:deploy
pbrew app:post-deploy
```

You can also have PBREW install dependencies with `pbrew app:install-deps`.

### Application Shell
When you want to interact with your application you should use `pbrew app:sh`. This will create a shell with all the needed environment variables, such as `PLATFORM_RELATIONSHIPS`.

### Database
You can use `pbrew db:sql` to access the database shell. The same database service is shared across all your projects to save system resources. Keep in mind that your databases will be prefixed with the project name, which is just the name of the directory your project is in.

I.e. if your project is in the directory `contextualcode` then all the databases for that project will be `contextualcode_<name>`.

You can also make a SQL dump of a database with `pbrew db:dump`.

### Stop Project(s)
You can stop a project with `pbrew p:stop`. This will stop only the services that project is using and only if those services aren't being used by another project. If you have two projects both using a database then you would have to stop both projects for the database service to also stop.
You can stop all projects with `pbrew all:stop`.

### Set Variables
You can set variables for a project with `pbrew var:set <key> <value>`. You can also use the `-g` option to set it globally for all projects.

```
pbrew var:set -g <key> <value>
```

You can view all the set variables with `pbrew var:list`.

### Router
You can start, stop, add projects, and remove projects from the main router.

```
pbrew router:start
pbrew router:stop
pbrew router:add
pbrew router:del
pbrew router:list
```


## Config
You can configure PBREW by adding a config.yaml file to PBREW's root application directory.

# Shell
You can configure the shell PBREW uses when you enter an application shell.
```
shell: bash
```

### Service Overrides
You can define custom service mappings that bypass PBREW's service handler. This can be used to override an existing service that PBREW already supports or to add support for a new service.

Example...
```
service_overrides:
    -
        type: "solr:*"
        host: example.com
        path: solr/core
        port: 8983
        scheme: solr
```
This would override the SOLR service for PBREW.


## TODO
- expansion/completion of nginx config generator

### Things that might be implemented
- app dependencies
- support for other languages (Go, Python, etc)
- workers
- cron jobs

### Things that don't work
- anything that relies on the app being in the /app directory...please use the PLATFORM_DIR environment variable
- app.web.commands.start ignored

### Things that won't be implemented
- varnish, all routes that point to varnish are passed through to the app