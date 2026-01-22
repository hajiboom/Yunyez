CC=gcc
CFLAGS=-Wall -Wextra -std=c99 -g
LDFLAGS=
TARGET=virtual_capture
SOURCES=virtual_capture.c

all: $(TARGET)

$(TARGET): $(SOURCES)
	$(CC) $(CFLAGS) -o $(TARGET) $(SOURCES) $(LDFLAGS)

clean:
	rm -f $(TARGET)

install:
	sudo cp $(TARGET) /usr/local/bin/

.PHONY: all clean install