build:
	@go build -o YapPad .

run: build
	./YapPad
