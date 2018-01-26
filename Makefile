build-linux-amd64:
	mkdir -p dist/goscatter-linux-amd64
	GOOS=linux GOARCH=amd64 go build -o ./dist/goscatter-linux-amd64/goscatter ./_main
	cp ./_main/conf.json ./dist/goscatter-linux-amd64
	zip -r ./dist/goscatter-linux-amd64.zip ./dist/goscatter-linux-amd64/
	rm -rf ./dist/goscatter-linux-amd64
