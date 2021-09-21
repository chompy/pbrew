# TODO
- list running projects to be able to determine what services can be stopped on project:stop
- expansion/completion of nginx config generator

# Things that might be implemented
- app dependencies
- solr
- support for other languages
- workers
- cron jobs

# Things that don't work
- anything that relies on the app being in the /app directory...please use the PLATFORM_DIR environment variable
- app.web.commands.start ignored

# Things that won't be implemented
- varnish, all routes that point to varnish are passed through to the app