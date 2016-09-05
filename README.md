Project Status: Complete | Adding features when needed
# k8svirt
A simple reverse proxy which emulates the behavior of a k8s cluster.

# usage
```
k8svirt [application dir]
```

## application directory
Application directory should contain one or more service directories.

## service directory
Directory which contains a web service. Must contain a `k8svirt-config.json` file to specify behavior to k8svirt.

# `k8svirt-config.json`
A `k8svirt-config.json` file specifies a service's behavior for k8svirt.  

The following keys are required to be present in the `k8svirt-config.json` file:
- `path`
	- Path service can be accessed by via the k8svirt proxy
	- Example: `/v1/auth/google`
- `server`
	- The url service will be running on
	- Must be a proper url
		- *GOOD* `http://127.0.0.1:3000`
		- *BAD* `127.0.0.1:3000`

# building
```
make build
```
