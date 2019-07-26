# faas

## Try it

You can make (almost) any exported Go function into a API using the syntax:

```
/?import=IMPORTPATH&func=FUNCNAME(...)
```

Because of how URLs are handled, you do need to URL encode the strings so that `FUNCNAME("hello, world") -> FUNCNAME(%22hello,%20world%22)`.

- IngredientsFromURL: https://faas.schollz.com/?import=github.com/schollz/ingredients&func=IngredientsFromURL(%22https://cooking.nytimes.com/recipes/12320-apple-pie%22)
- Md5Sum: https://faas.schollz.com/?import=github.com/schollz/utils&func=Md5Sum(%22hello,%20world%22)

## Get started

You need to [install Docker](https://docs.docker.com/install/linux/docker-ce/ubuntu/#install-docker-engine---community-1), and make sure `gzip` is installed.

Then install with Go:

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

The first time you run it will take a minute to build the container, after which it will save the container and load after the container times out.

## License

MIT
