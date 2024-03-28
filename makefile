TARGET := "pokedexcli"

run: build
	bin/$(TARGET)

build:
	go build -o bin/$(TARGET)

clean:
	$(RM) bin/*
