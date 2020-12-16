default:
	rm -rf dist
	mkdir dist
	GOARCH=arm go build -o ./dist/btc-price ./cmd/btc-price
	cp ./assets/7in5_V2.png ./dist/
	scp ./dist/* 10.1.1.231:~/btc-price/