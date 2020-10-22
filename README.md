# Go Server
A simple Go HTTP server that simply returns HTTP status codes and server time or errors in a JSON format. 
Contains the ability to be built as a Docker image and run inside of a Docker container with error and success messages being submitted to Loggly.

Recommended Docker build and run procedure (clutterless):

	- $ docker build -t server --rm --quiet .	 						// Builds quietly 
	- $ docker run -p 8081:8000/tcp --env-file env.list -d --rm --name aserver server		// Runs detached, auto removal.
	- $ docker logs aserver -f									// Live container output
	
Process for removing container and images:

	- $ docker stop aserver						// Stop container myagent
	- $ docker rmi $(docker images -a -q)				// Remove all stopped images
	
Changelog:
-------------------------------
[10/1/2020]: 

	- Initial server created with full deployment functionality including ability to run inside of Docker.

[10/2/2020]: 

	- Fixed: Implemented how HTTP status codes were returned when an invalid request was made (insures proper browser behavior).
	- Tested: Ran inside of Docker on an EC2 instance, verified proper messages were sent to Loggly and web browser.
