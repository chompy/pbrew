package core

/*func GetUpstreamHost(proj *project.Project, upstream string, allowServices bool) (string, error) {
	upstreamSplit := strings.Split(upstream, ":")
	// itterate apps and services to find name match
	// TODO this should use relationships but those only get resolved when
	// services are opened...sooo??
	for _, app := range proj.Apps {
		if app.Name == upstreamSplit[0] {
			return proj.GetDefinitionHostName(app), nil
		}
	}
	for _, serv := range proj.Services {
		if serv.Name == upstreamSplit[0] {
			// forward to app if allowServices is false
			if !allowServices {
				for _, relationship := range serv.Relationships {
					rlSplit := strings.Split(relationship, ":")
					return GetUpstreamHost(proj, fmt.Sprintf("%s:http", rlSplit[0]), allowServices)
				}
			}
			// TODO use relationship to determine port
			port := 80
			switch serv.GetTypeName() {
			case "varnish", "solr":
				{
					port = 8080
					break
				}
			}
			return fmt.Sprintf("%s:%d", proj.GetDefinitionHostName(serv), port), nil
		}
	}
	return "", errors.Wrapf(ErrUpstreamNotFound, "upstream %s not found", upstream)
}
*/
