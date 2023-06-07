.PHONY: build run clean


# Build the executable
build:
	go build -o scorecard4purl scorecard4purl.go

# Run the program with a PURL
run:
	go run scorecard4purl.go $(PURL)

# Clean the generated executable
clean:
	rm -f scorecard4purl
