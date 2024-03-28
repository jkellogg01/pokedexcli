TARGET := "pokedexcli"

dev: build
	bin/$(TARGET) -v

run: build
	bin/$(TARGET)

build:
	go build -o bin/$(TARGET)

clean:
	$(RM) bin/*
