# TODO
- project database for tracking
    - assign and map ports
    - list running projects to be able to determine what services can be stopped on project:stop
- auto install php extensions
- build hooks
- router port 80/443 option (needs sudo)
- router ssl
- maybe...
    - dependencies
    - solr
    - support for other languages

# Things that don't work
- anything that relies on the app being in the /app directory...please use the PLATFORM_DIR environment variable
- app.web.commands.start ignored