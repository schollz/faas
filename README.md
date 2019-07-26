# faas

## Get started

You need to [install Docker](https://docs.docker.com/install/linux/docker-ce/ubuntu/#install-docker-engine---community-1), and make sure `gzip` is installed.

Then install with Go:

```
go get -v github.com/schollz/faas
```

Now you can run:

```
faas --debug
```

Now you can try it out:

```
curl http://localhost:8090/?import=github.com/schollz/utils&func=Md5Sum(%22hello,%20world%22)
```

The first time you run it will take a minute to build the container, after which it will save the container and load after the container times out.

## License

MIT