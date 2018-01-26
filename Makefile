build-linux-amd64:
	mkdir -p dist/goscatter-linux-amd64
	GOOS=linux GOARCH=amd64 go build -o ./dist/goscatter-linux-amd64/goscatter ./_main
	cp ./_main/conf.json ./dist/goscatter-linux-amd64
	cd ./dist; zip -r ./goscatter-linux-amd64.zip ./goscatter-linux-amd64/
	rm -rf ./dist/goscatter-linux-amd64

clean:
	rm -rf ./dist
