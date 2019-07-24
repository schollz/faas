# faasss

## Idea

This is a `FaassS` - *functions as a stupidly simple service*. It's [iron.io/functions](https://github.com/iron-io/functions), or [zeit/now](https://github.com/zeit/now-cli) or [openfaas](https://github.com/openfaas/faas) but more simple and more stupid.

What does it do? You can run any Go package function as a service. 

**Why is it simple?** Zero coding, zero config. Just make a HTTP request with the name of the package, and the name of the function, and that's it!

**Why is it stupid?** It doesn't do anything fancy.

## Get started

No coding required! Make a FaassS from some code that already exists! For instance, lets make a service that returns ingredients in a recipe URL using the package `github.com/schollz/ingredients`.

```bash
$ curl https://faas.schollz.com/?package=github.com/schollz/ingredients&func=ParseIngredients
Your function is at: https://faas.schollz.com/3b302a
```

Get data from it:

```bash
$ curl -d '{"hello":"world"}' \ 
	-H "Content-Type: application/json" \
	-X POST https://faas.schollz.com/3b302a
{"out": "world"}
```

## How does it work?

That command will parse the package `github.com/schollz/googleit` and check the `Run(..)` function which may look like anything, e.g.

```go
func Run(hello string) (out string, err error) {
	out = hello
	return
}
```

It will then write a Go server that runs that function using a `POST` request with the parameters and returning a JSON with the results of the function. It will then use Docker to run a container of this server and assign a route `/<id>` where you can `POST` JSON content to.

```bash
$ curl -d '{"hello":"world"}' \ 
	-H "Content-Type: application/json" \
	-X POST https://faas.schollz.com/3b302a
{"out": "world"}
```


## Todo

Parse Golang file, determine function. Use AST to gather the parameters and the results.

See https://play.golang.org/p/cG8sDVK0YSU (archived: https://share.schollz.com/2jyio9/main.go)

Using the source code, generate a struct for the input and the output.

```
type Input struct {
	FirstParam string `json:"firstparam"`
	SecondParam  []string `json:"secondparam"`
}
```

```
type Output struct {
	EverythingButErr string
}
```

## List current images

```
# list images
docker container ls --format "{{.Image}}"
```

Look for images with prefix `faas/`


## Create new image

`GET https://faas.schollz.com/new/?url=https://share.schollz.com/1/asldfkjasd&key=alskdfj`

Uses handler from to 3rd party (Github, share.schollz.com, etc.)

## Build new image

The build name and running name is the sha256sum of the data in the file: faas/ID

```
docker kill faas/ID
docker build -t faas/ID .
docker run -d -p AVAILABLEPORT:8080 --name faas/ID faas/ID
docker image prune -a -f
```

## Route iamge

Routing the image and accessing that function is then available at 

```
https://faas.schollz.com/ID/.
```

The `faas` will trim the ID prefix and route to the corresponding local port containing the function.
