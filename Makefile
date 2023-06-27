NAME=airquality
BINDIR=bin
SOURCES=$(wildcard *.go) 
SENSIRON_SOURCES=sen5x_i2c.c sensirion_common.c sensirion_config.h sensirion_i2c.h sensirion_i2c_hal.h sen5x_i2c.h sensirion_common.h sensirion_i2c.c sensirion_i2c_hal.c

all: ${BINDIR} ${BINDIR}/${NAME}

${SENSIRON_SOURCES}:
	curl -s -O -L https://raw.githubusercontent.com/Sensirion/raspberry-pi-i2c-sen5x/master/$@
	sed -i -e "s+/dev/i2c-1+/dev/i2c-0+" $@

${BINDIR}:
	mkdir -p ${BINDIR}
	
${BINDIR}/${NAME}: ${SENSIRON_SOURCES}  ${SOURCES} 
	go build -o $@

run:
	go run ${SOURCES}

clean:
	@go clean
	-@rm -rf ${BINDIR} ${SENSIRON_SOURCES} 2>/dev/null || true
