FROM golang:1.18.0

LABEL MAINTAINER="Ivan Izaguirre <ivanizag@gmail.com>"

RUN apt-get update
RUN apt-get install -y libsdl2-dev mingw-w64 unzip

RUN wget https://www.libsdl.org/release/SDL2-devel-2.0.12-mingw.tar.gz
RUN tar -xzf SDL2-devel-2.0.12-mingw.tar.gz
RUN cp -r SDL2-2.0.12/x86_64-w64-mingw32 /usr

RUN wget https://www.libsdl.org/release/SDL2-2.0.12-win32-x64.zip
RUN unzip SDL2-2.0.12-win32-x64.zip -d /sdl2runtime

COPY buildindocker.sh .
RUN chmod +x buildindocker.sh

CMD ["./buildindocker.sh"]
