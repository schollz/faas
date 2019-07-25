# faas

```
cd $GOPATH/src/github.com/schollz/faas/pkg/gofaas
go test -v -cover
docker run -d -p 8081:8080 -t faas-1
curl http://localhost:8081/?import=github.com/schollz/ingredients&func=NewFromURL(%22https://www.allrecipes.com/recipe/10813/best-chocolate-chip-cookies%22)
```
