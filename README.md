# faas

Make any (Go) function into an API with one HTTP request.

This is a FaaS - *functions as a service* but its more of a FaaSSS - *functions as a stupidly simple service*. It's [iron.io/functions](https://github.com/iron-io/functions), or [zeit/now](https://github.com/zeit/now-cli) or [openfaas](https://github.com/openfaas/faas) or [apex](https://github.com/apex/apex) or [sky-island](https://github.com/briandowns/sky-island) but more simple and more stupid. There is no coding, no init-ing, no pushing, no updating, no bumping. You Just make a HTTP request with the name of the package, the name of the function, and any input. Right now it only works for Go. 


## Examples 

You can make (almost) any exported Go function into a API! 

Run [`Md5Sum`](https://github.com/schollz/utils/blob/adaa47085f7b6b1c3e1ecfebfb18028e08e0bde2/hash.go#L29-L34) to get a md5 hash of "hello, world":

```bash
$ curl https://faas.schollz.com/?import=github.com/schollz/utils&func=Md5Sum(%22hello,%20world%22)
e4d7f1b4ed2e42d15898f4b27b019da4
```

Run [`IngredientsFromURL`](https://github.com/schollz/ingredients/blob/23a2a0c2d9dc8988c33acf7650ae9284a59d0b20/ingredients.go#L153-L160) to get the ingredients from any website:

```bash
$ curl https://faas.schollz.com/?import=github.com/schollz/ingredients&func=IngredientsFromURL(%22https://cooking.nytimes.com/recipes/12320-apple-pie%22)
{"ingredients": [{"name":"butter","comment":"unsalted","measure":{"amount":2,"name":"tablespoons","cups":0.25}},{"name":"apples","measure":{"amount":2.5,"name":"pounds","cups":14.901182654402104}},{"name":"allspice","comment":"ground","measure":{"amount":0.25,"name":"teaspoon","cups":0.0104165}},{"name":"cinnamon","comment":"ground","measure":{"amount":0.5,"name":"teaspoon","cups":0.020833}},{"name":"salt","comment":"kosher","measure":{"amount":0.25,"name":"teaspoon","cups":0.0104165}},{"name":"sugar","comment":"plus 1 tablespoon","measure":{"amount":0.75,"name":"cup","cups":1.5}},{"name":"flour","comment":"all purpose","measure":{"amount":2,"name":"tablespoons","cups":0.25}},{"name":"cornstarch","measure":{"amount":2,"name":"teaspoons","cups":0.083332}},{"name":"apple cider vinegar","measure":{"amount":1,"name":"tablespoon","cups":0.125}},{"name":"pie dough","measure":{"amount":1,"name":"whole","cups":0}},{"name":"egg","measure":{"amount":1,"name":"whole","cups":0}}],"err": null
```

Run `MarkdownToHTML` from a [Go playground](https://play.golang.org/p/9xzE8Ivwupk):

```bash
$ curl -d '{"markdown":"*hello*,**world**"}' \
 	-H "Content-Type: application/json" -X POST \
 	https://faas.schollz.com/?import=https://play.golang.org/p/9xzE8Ivwupk.go&func=MarkdownToHTML
<p><em>hello</em>,<strong>world</strong></p>
```

## Usage 

You can use `GET` or `POST` to submit jobs.

For the `GET` requests the syntax is

```
/?import=IMPORTPATH&func=FUNCNAME(param1,param2...)
```

The `IMPORTPATH` is the import path (e.g. github.com/x/y) or a URL containing the file with the function. The `FUNCNAME` is the name of the function. Note, you do need to URL encode the strings so that `FUNCNAME("hello, world") -> FUNCNAME(%22hello,%20world%22)`


For the `POST` requests the syntax is:

```
/?import=IMPORTPATH&func=FUNCNAME
```

with the body with the inputs `{"param":"value"}`.

That's it! The first time you run it will take ~1 minute while the Docker image is built.


## How does it work?

When you make a `POST`/`GET` request to the `faas` server it will locate the given function and given package and it will generate a Docker container that accepts a JSON input with the parameters and outputs a JSON containing the output variables. This Docker container automatically shuts down after some time of inactivity. Subsequent requests then will load the previously built container for use.

## Host yourself

You need to [install Docker](https://docs.docker.com/install/linux/docker-ce/ubuntu/#install-docker-engine---community-1), and make sure `gzip` is installed.

Then build `faas` with Go:

```
git clone https://github.com/schollz/faas
cd faas
go generate
go build -v
```

Now you can run:

```
./faas --debug
```

Now you can try it out:

```
curl http://localhost:8090/?import=github.com/schollz/utils&func=Md5Sum(%22hello,%20world%22)
```

OR post data:

```
curl -d '{"s":"hello, world"}' -H "Content-Type: application/json" -X POST http://localhost:8090/?import=github.com/schollz/utils&func=Md5Sum
```

Note that the JSON `"s"` comes from the function `Md5Sum` itself.

## License

MIT
