# k8eraid
(KAY-ter-ade)

[![Travis](https://api.travis-ci.com/bloomberg/k8eraid.svg?branch=master)](https://travis-ci.com/bloomberg/k8eraid)

## What does this do?

The point of k8eraid is to provide a relatively simple, unified method for reporting on Kubernetes resource issues.

How is this different from metrics based alerters (ie: Prometheus with AlertManager)?

This tool directly integrates with the Kubernetes API to look up the actual real state of resources and their changes. That means where metrics based systems are good for looking up a snapshot of status based on specifications (X number of pods running right now), k8eraid can look up the current status as well as see if the status has recently changed by looking at the timestamps for API metadata like deletion times and pod state readiness times. Additionally, the configuration for k8eraid is simple enough that it is easy to expand or remove alerting and monitoring as you see fit.  Not only that but k8eraid is incredibly lightweight, using about 100KB of memory at runtime to look up the status of 100 individual resources, making it a prime, lightweight candidate for allowing multiple teams in a shared environment to deploy per-namespace and allowing them to manage alerting as they see fit!

## How does it work?

- Create a JSON file specifying what resources you want to monitor, and how you want to alert.
- Make the config into a configmap.
- Make an appropriate role with permissions for k8eraid to read all resources in the Apps and Core APIs.
- Deploy it to your cluster with k8eraid.
- Get annoyed at the fact that you now get alerts when things go wrong in your cluster.

## Which Kubernetes versions are supported?

Kubernetes version | Works
-------------------|------
1.6.X              | :white_check_mark:
1.7.X              | :white_check_mark:
1.9.X              | :white_check_mark:
1.10.X             | :white_check_mark:

## What does k8eraid monitor?

Resource    | Statuses
----------- | -------------------
Pods	    | Minimum pod count, pod restarts, Failed scheduling, Stuck terminating
Deployments | Minimum replica count
Daemonsets  | Minimum replica count, Failed scheduling
Nodes       | Out of disk, Memory pressure, Disk pressure, Node readiness, Node count

K8eraid can not only perform these checks against single resources, but you can specify "global" rules using "*".  Additionally, global rules can use filters based on resource labels!

## What alert methods does k8eraid support?

Alert type  | Options
------------|---------
stderr      |
smtp	    | Mail server, Port, Password ENV var, Subject, From address, To address
pagerdutyV2 | Service key ENV var, Proxy server, Subject
webhook     | Server, Proxy server, Subject

## Get it from [DockerHub](https://hub.docker.com/r/bloomberg/k8eraid):

```sh
# get from DockerHub
docker pull bloomberg/k8eraid
```



## Awesome! So how does configuration work?

There are five types of objects in a config- "deployments", "pods", "daemonsets", "nodes", and "alerters". Each of these objects contain one or more desired definitions. There are a few important rules that you will need to remember when configuring your rules, most of these are due to the way the kubernetes client functions in `list` vs `get` functions.

- The config is self-reloading. You do not need to redeploy k8eraid when you update the configmap.
- If using a wildcard for a POD, you MUST specify a valid filterLabel
- If specifying a name for any target resource, you MUST specify a valid filterNamespace
- If your pendingThreshold is too short for a POD rule, you may get alerts for normal pod startups.
- For DEPLOYMENT and DAEMONSET type resources- "filter" can either be a literal string for a namespace, or a key/value pair string for a metadata label

### Pod configuration examples

- Check for pod restarts and failures scheduling of pod named "foobarbaz-pod" in the "default" namespace. But only if the pod has existed in kubernetes for at least 120 seconds. Send errors to stderr
``` json

{
	"name": "foobarbaz-pod",
	"filterNamespace": "default",
	"filterLabel": "",
	"alerter": "stderr",
	"reportStatus": {
		"minPods": 1,
		"podRestarts": true,
		"failedScheduling": true,
		"pendingThreshold": 120
	}
}

```

- Check for failures scheduling or deleting ALL pods, in any namespace with the metadata label "monitor=true", but only if the pod has existed in kubernetes for at least 10 seconds. Send email using "example-email" alerter
``` json

{
	"name": "*",
	"filterNamespace": "",
	"filterLabel": "monitor=true",
	"alerter": "example-email",
	"reportStatus": {
		"minPods": 1,
		"podRestarts": false,
		"failedScheduling": true,
		"stuckTerminating": true,
		"pendingThreshold": 10
	}
}

```

### Deployment configuration examples

- Check to make sure the "foobar-deployment" deployment has at least 3 ready pods, but only if "foobar-deployment" has been around for at least 10 seconds. Use pagerduty to send an alert.
``` json

{
	"name": "foobar-deployment",
	"filter": "default",
	"alerter": "example-pagerduty",
	"reportStatus": {
		"minReplicas": 3,
		"pendingThreshold": 10
	}
}

```

- Check to make sure all deployments older than 30 seconds have at least 1 ready pod. Send alerts to stderr.
``` json

{
	"name": "*",
	"filter": "",
	"alerter": "stderr",
	"reportStatus": {
		"minReplicas": 1,
		"pendingThreshold": 30
	}
}

```

### Daemonset configuration examples

- Check to see if the daemonset "daemon-of-glory" has the expected number of replicas deployed, checking for failed scheduling- assuming the Daemonset is at least 10 seconds old. Send alerts to stderr.
``` json

{
	"name": "daemon-of-glory",
	"filter": "",
	"alerter": "stderr",
	"reportStatus": {
		"checkReplicas": true,
		"failedScheduling": true,
		"pendingThreshold": 10
	}
}

```

### Node configuration examples

- Examine all nodes with the label "monitor=true" that are at least 5 minutes old. Check to make sure there are at least 10 nodes in the cluster, and watch for OutOfDisk, MemoryPressure, DiskPressure, and Readiness issues. Send alerts to stderr.
``` json

{
	"name": "*",
	"filter": "monitor=true",
	"alerter": "stderr",
	"reportStatus": {
		"minNodes": 10,
		"outOfDisk": true,
		"memoryPressure": true,
		"diskPressure": true,
		"readiness": true,
		"pendingThreshold": 300
	}
}

```

### Alerter configuration

- stdout is a default constant alerter name that will always spew errors to stdout where the application is running. No special configuration is needed.
- All other alert types may be configured multiple different ways each with unique names- allowing you to change alert behavior based on your rules as desired.

- Example smtp alert named "example-email", this will email me@example.com when called upon
``` json

{
	"name": "example-email",
	"toAddress": "me@example.com",
	"fromAddress": "kubernetes@example.com",
	"mailServer": "smtp.example.com",
	"port": 25,
	"subject": "Observed issue with Kubernetes cluster"
}

```

- Example smtp alert named "gmail-email", this will email me@example.com from foo@gmail.com when called upon, using the password saved in GMAIL_PW ENV var for authentication.
``` json

{
	"name": "gmail-email",
	"toAddress": "me@example.com",
	"fromAddress": "foo@gmail.com",
	"mailServer": "smtp.gmail.com",
	"port": 587,
	"passwordEnvVar": "GMAIL_PW",
	"subject": "Observed issue with Kubernetes cluster"
}

```

- Example Pagerduty alert name "example-pagerduty", this will trigger a pagerduty alert using the value of the injected ENV variable of PD_KEY as the service key, using http://proxy.example.com:80 as an http proxy.
``` json

{
	"name": "example-pagerduty",
	"serviceKeyEnvVar": "PD_KEY",
	"proxyServer": "http://proxy.example.com:80",
	"subject": "Observed issue with Kubernetes cluster"
}

```

## Contributing

Got features or bugfixes? please feel free to contribute with code or issues!

### Required tools
- Docker
- [`gofmt`](https://golang.org/cmd/gofmt/)
- [`golint`](https://github.com/golang/lint)

### Useful editor integrations
- Vim: [`vim-go`](https://github.com/fatih/vim-go)
- Emacs [`go-mode.el`](https://github.com/dominikh/go-mode.el)
- VSCode [`vscode-go`](https://github.com/Microsoft/vscode-go)

### Building binary with valid Golang environment

`make build`

### Building binary with valid Docker installation

`make buildcontainer`

### Testing code with a valid Docker installation

`make testcontainer`

### Building a Docker container

`make container`
