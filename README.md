# Go Server
A simple Go HTTP server that simply returns HTTP status codes and server time or errors in a JSON format. 
Contains the ability to be built as a Docker image and run inside of a Docker container with error and success messages being submitted to Loggly.

Recommended Docker build and run procedure (clutterless):

	- $ docker build -t server --rm --quiet .	 						// Builds quietly and output hash once done.
	- $ docker run -p 8081:8080/tcp --env-file env.list -d --rm --name aserver asingh2-rates-api	// Runs detached, auto removal when finished.
	- $ docker logs aserver -f									// Live container output
	
Process for removing container and images:

	- $ docker stop aserver						// Stop container myagent
	- $ docker rmi $(docker images -a -q)				// Remove all stopped images
	
Process for saving Docker image as .tar file and loading into Docker.
	
	- $ docker save -o /Documents/agent.tar agent:latest		// Saves the agent image in Documents directory as agent.tar
	- $ docker load -i /Documents/agent.tar				// Loads the agent.tar from Documents directory into Docker.
	
Changelog:
-------------------------------
[10/1/2020]: 

	- Initial server created with full deployment functionality including ability to run inside of Docker.

[10/2/2020]: 

	- Fixed: How HTTP status codes were returned when an invalid request was made (insures proper browser behavior).
	- Tested: Ran inside of Docker on an EC2 instance, verified proper messages were sent to Loggly and web browser.

[10/22/2020]: 

	- Updated: Status endpoint to return DynamoDB table name and number of item entries. 
	- Implemented: DynamoDB functionality and error messages for Loggly.
	- Implemented: Created a "asingh2/all" endpoint to return all data from a table encoded as JSON.
	- Implemented: A "forbidden" function to handle invalid paths, return a HTTP 404 status and submit all possible attempts to Loggly.
	- Updated: Dockerfile file to include multi-stage build process to shrink the size of final Docker image by nearly 99% (1.5GB to 19MB).
	- Updated: Default listening port from 8000 to 8080/tcp.
	- Improved: Regex that handles invalid endpoint requests to include more invalid paths for logging purposes.
	
[10/30/2020]:

	- Implemented: A new "search" endpoint to query specific data from the API endpoint.
	- Implemented: Input sanitization of all queries that are made to the endpoint using a package in Go.
	- Implemented: Input validation and sanitization on top of using Go package to hopefully mitigate database attack vectors. 
	- Improved: When Loggly messages were reported to make it less redundant and reflect what's actually happenning and when.
	- Improved: Loggly message levels to reflect errors, warnings, and general information for better troubleshooting in Loggly. 
	- Improved: Formatting of server.go code using "go fmt" command.
	
[11/18/2020]:

	- Updated: Location where Docker pulled required images from (DockerHub -> ECR).
	- Added: "buildspec.yml" file for auto deployment using AWS CodePipeline.
