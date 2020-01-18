# REST
Go library to create your own REST services easily

# Examples

## Simple Server

There are some _examples_ into the folder with the same name. The
most basic one is the simple server. You can find in under
_simple_ subfolder.

Update example dependencies with

```
$ go get -t ./examples/simple/...
```

and run it with:

```
$ go run ./examples/simple/server/main.go
Jan 26 20:41:31.880 I 01 example | Server unix socket started at /tmp/example_940156266/unix.socket
Jan 26 20:41:31.880 I 02 example | Server HTTPS started in port 8443
```

once it is up, you can reach its API by using any browser, curl or any other manner
to call a REST API. IN this simple exapple you can create resources, list the existing
ones, list or delete a specific one. Reources can include whatever value that you prefer.

Next are the available operations:
| Method | URL | Description |
|--------|-----|-------------|
| POST | /1.0/resources | creates a new resource |
| GET | /1.0/resources | list all the available resources |
| GET | /1.0/resource/[id] | returns the details of a created resource |
| PUT | /1.0/resource/[id] | updates the value of the resource with [id] identifier |
| DELETE | /1.0/resource/[id] | deletes the [id] resource |

You can request an operation using the available unix socket (look at the output 
traces when service starts up to know the exact path to the socket)
```
curl --unix-socket /tmp/example_940156266/unix.socket s/1.0/resources
```
or by requesting the HTTPS endpoint
```
curl -k https://localhost:8443/1.0/resources
```

### Build docker container

You can deploy simple server in a docker container. There is a Makefile
ready to build such container in just one step:

```
make
```

that would create a docker container for testing and building source code
and a second definitive one taking the binary and starting it up in 8443 port.
The result is an image tagged with *simple-server:[TAG]*. The tag value is an
abbreviation of the last git commit.

You can run locally a container based on such image with:

```
docker run -p 8443:8443 simple-server
```

Now, the port 8443 is accessible in localhost to reach the API

### Deploy to Kubernetes

A further step is deploy this simple server as a [Kubernetes](https://kubernetes.io/) pod.
In the root of the project you can find a deployment file named _simple-server-deployment.yaml_
to accomplish this.

First of all, let's publish our image to a repository. For this example
we use [docker hub](https://hub.docker.com/), but you can use any other.

Login to hub, tag the image and push.

```
docker login --username [username] --password [password]
docker tag [image-id] [username]/[repo]:[tag]
docker push [username]/[repo]
```

!!! Note:
    Replace the placeholders (username, password, image-id, repo, tag) by their value


Next, deploy to the cluster. You can use [Minikube](https://minikube.io) to test it locally.
Asuming that it is installed, start it if needed

```
minikube start
```

Deploy:

```
kubectl create -f simple-server-deployment.yaml
kubectl get pods
```

!!! Note:
    Edit the file and replace the image path with the one created in the docker hub

Look for the exposed url
```
$ minikube service simple-server --url
https://192.168.39.200:32028
```

Now you know where is the endpoint where REST API is exposed:

```
curl -k https://1192.168.39.200:32028/1.0/resources
```