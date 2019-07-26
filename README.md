# faas

This is a FaaS - *functions as a service* but its more of a FaaSSS - *functions as a stupidly simple service*. It's [iron.io/functions](https://github.com/iron-io/functions), or [zeit/now](https://github.com/zeit/now-cli) or [openfaas](https://github.com/openfaas/faas) but more simple and more stupid. There is no coding, no init-ing, no pushing, no updating, no bumping. You Just make a HTTP request with the name of the package, the name of the function, and any input.

You can make (almost) any exported Go function into a API. Some examples:


- IngredientsFromURL: https://faas.schollz.com/?import=github.com/schollz/ingredients&func=IngredientsFromURL(%22https://cooking.nytimes.com/recipes/12320-apple-pie%22)
- Md5Sum: https://faas.schollz.com/?import=github.com/schollz/utils&func=Md5Sum(%22hello,%20world%22)


The syntax is simply:

```
/?import=IMPORTPATH&func=FUNCNAME(...)
```

That's it! Because of how URLs are handled, you do need to URL encode the strings so that `FUNCNAME("hello, world") -> FUNCNAME(%22hello,%20world%22)`. Also, the first time you run it will take ~1 minute while the Docker image is built.


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
