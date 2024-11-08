### TODO list

#### Done
- [x] Routing
	- [x] URL Validation
	- [x] URL routing for buckets
	- [x] URL routing for objects
- [x] Buckets handling
	- [X] Implement create the bucket  endpoint
		- [x] Create the empty directory related to bucket
	- [x] Logging
	- [x] Implement listing all buckets endpoint
	- [x] Implement deleting the bucket endpoint
		- [x] Delete the empty directory related to bucket
	- [x] Logging
	- [x] Bucket status handling
	- [x] Modified time update
- [x] Objects handling
	- [x] Upload a new object
	- [x] Retrieve an object
	- [x] Delete an object
	- [x] Logging
	- [x] Modified time update
	- [x] Unit Manual Testing
- [x] Finish line
	- [x] Change responses format to XML instead of raw text
	- [x] Appropriate response headers
	- [x] Help flag
	- [x] Full manual testing (SKIPPED THE TRAVERSE DIRECTORY VULNERABILITY)
	- [x] Check correspondence with project requirements

#### Future backlog
- [ ] Path traversal vulnerability
- [ ] Incorrect flag handling
- [ ] Bucket get endpoint
- [ ] Refactor
	- [ ] csv headers line add and its appropriate handling
	- [ ] Implement all REST constraints 
		- [ ] Options for each resource
		- [ ] Links for each available resource in initial uri
		- [ ] MVC Pattern
		... other constraints of REST
- [ ] Different error response codes, more clear error messages
- [ ] SOAP constraints (???)
- [ ] Full testing covering
- [ ] Authorization
- [ ] Front-end 
- [ ] Containerize the application
- [ ] Ansible, Yaml configuration file
- [ ] CI/CD

