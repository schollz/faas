# faas

This is a FaaS - *functions as a service* but its more of a FaaSSS - *functions as a stupidly simple service*. It's [iron.io/functions](https://github.com/iron-io/functions), or [zeit/now](https://github.com/zeit/now-cli) or [openfaas](https://github.com/openfaas/faas) or [apex](https://github.com/apex/apex) or [sky-island](https://github.com/briandowns/sky-island) but more simple and more stupid. There is no coding, no init-ing, no pushing, no updating, no bumping. You Just make a HTTP request with the name of the package, the name of the function, and any input. Right now it only works for Go. You can make (almost) any exported Go function into a API using `GET` or `POST` queries.


### Examples 

IngredientsFromURL(url) from [schollz/ingredients](https://github.com/schollz/ingredients): [/?import=github.com/schollz/ingredients&func=IngredientsFromURL("https://cooking.nytimes.com/recipes/12320-apple-pie")](https://faas.schollz.com/?import=github.com/schollz/ingredients&func=IngredientsFromURL(%22https://cooking.nytimes.com/recipes/12320-apple-pie%22)


Md5Sum(s) from [schollz/utils](https://github.com/schollz/utils): [/?import=github.com/schollz/utils&func=Md5Sum("hello, world")](https://faas.schollz.com/?import=github.com/schollz/utils&func=Md5Sum(%22hello,%20world%22))



Search(url) from [schollz/googleit](https://github.com/schollz/googleit):

```
$ curl -d '{"query":"mint chocolate chip cookie recipe","ops":{"NumPages":3,"MustInclude":["chocolate","chip","cookie","mint"]}}' -H "Content-Type: application/json" -X POST https://faas.schollz.com/?import=github.com/schollz/googleit&func=Search`
```

- MarkdownToHTML from [Go playground](https://play.golang.org/p/9xzE8Ivwupk):

```
$ curl -d '{"markdown":"*hello*,**world**"}' -H "Content-Type: application/json" -X POST https://faas.schollz.com/?import=https://play.golang.org/p/9xzE8Ivwupk.go&func=MarkdownToHTML
```

For the `GET` requests the syntax is

```
/?import=IMPORTPATH&func=FUNCNAME(param1,param2...)
```

Note, you do need to URL encode the strings so that `FUNCNAME("hello, world") -> FUNCNAME(%22hello,%20world%22)`


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
